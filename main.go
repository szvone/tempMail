package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alash3al/go-smtpsrv/v3"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

//go:embed static/*
var staticFS embed.FS

// Config 应用配置
type Config struct {
	AllowedDomains []string
	SMTPPort       string
	HTTPPort       string
	HTTPSPort      string
	CertFile       string
	KeyFile        string
	EnableHTTPS    bool
	StorageMode    string
}

// MailContent 邮件内容结构
type mailContent struct {
	from        string
	to          string
	title       string
	TextContent string
	HtmlContent string
	receivedAt  time.Time
}

var (
	config         Config
	storageFactory *StorageFactory
)

// 初始化配置
func initConfig() Config {
	cfg := Config{
		AllowedDomains: strings.Split(os.Getenv("ALLOWED_DOMAINS"), ","),
		SMTPPort:       getEnvOrDefault("SMTP_PORT", "25"),
		HTTPPort:       getEnvOrDefault("HTTP_PORT", "80"),
		HTTPSPort:      getEnvOrDefault("HTTPS_PORT", "443"),
		CertFile:       getEnvOrDefault("CERT_FILE", "./certs/server.pem"),
		KeyFile:        getEnvOrDefault("KEY_FILE", "./certs/server.key"),
		EnableHTTPS:    os.Getenv("ENABLE_HTTPS") == "true",
		StorageMode:    getEnvOrDefault("STORAGE_MODE", "memory"),
	}

	if len(cfg.AllowedDomains) == 0 || cfg.AllowedDomains[0] == "" {
		log.Fatal("错误：ALLOWED_DOMAINS 环境变量未设置")
	}

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func handler(c *smtpsrv.Context) error {
	to := strings.Trim(c.To().String(), "<>")
	from := strings.Trim(c.From().String(), "<>")
	msg, err := c.Parse()
	if err != nil {
		log.Printf("解析邮件失败: %v", err)
		return err
	}
	log.Printf("收到来自 %s 发送给 %s 的邮件", from, to)

	content := MailContent{
		From:        from,
		To:          to,
		Title:       msg.Subject,
		TextContent: msg.TextBody,
		HtmlContent: msg.HTMLBody,
		ReceivedAt:  time.Now(),
	}

	// 使用存储工厂存储邮件
	if storageFactory != nil {
		err := storageFactory.GetStorage().StoreMail(to, content)
		if err != nil {
			log.Printf("存储邮件失败: %v", err)
			return err
		}
	}

	log.Printf("来自 %s 发送给 %s 的邮件 邮件处理完毕！", from, to)
	return nil
}

func startSMTPServer() error {
	cfg := smtpsrv.ServerConfig{
		BannerDomain:    config.AllowedDomains[0],
		ListenAddr:      ":" + config.SMTPPort,
		MaxMessageBytes: 256 * 1024,
		Handler:         handler,
	}

	log.Printf("SMTP服务器正在启动于端口 %s...", config.SMTPPort)
	return smtpsrv.ListenAndServe(&cfg)
}

func startHTTPServer() {
	gin.SetMode(gin.ReleaseMode)
	httpSrv := gin.Default()

	// 添加恢复中间件
	httpSrv.Use(gin.Recovery())

	// 添加简单的访问日志
	httpSrv.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		log.Printf("[%s] %s %s %v", c.Request.Method, path, c.ClientIP(), time.Since(start))
	})

	setupRoutes(httpSrv)

	// 启动 HTTP 服务器
	go func() {
		log.Printf("HTTP服务器正在启动于端口 %s...", config.HTTPPort)
		if err := httpSrv.Run(":" + config.HTTPPort); err != nil {
			log.Printf("HTTP服务器启动失败: %v", err)
		}
	}()

	// 根据配置决定是否启动 HTTPS 服务器
	if config.EnableHTTPS {
		log.Printf("HTTPS服务器正在启动于端口 %s...", config.HTTPSPort)
		if err := httpSrv.RunTLS(":"+config.HTTPSPort, config.CertFile, config.KeyFile); err != nil {
			log.Printf("HTTPS服务器启动失败: %v", err)
		}
	}
}

func setupRoutes(r *gin.Engine) {
	// 使用嵌入的静态文件系统
	staticSubFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("无法加载静态文件: %v", err)
	}

	// 静态文件服务 - 使用嵌入的文件系统
	r.StaticFS("/static", http.FS(staticSubFS))

	// 主页服务
	r.GET("/", func(c *gin.Context) {
		data, err := staticFS.ReadFile("static/index.html")
		if err != nil {
			c.String(500, "无法加载主页")
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	})

	r.GET("/getAllowedDomains", func(c *gin.Context) {
		c.JSON(200, gin.H{"allowedDomains": config.AllowedDomains})
	})

	r.GET("/getMail/:randomString", handleGetMail)
}

func handleGetMail(c *gin.Context) {
	mailHead := c.Param("randomString")

	// 使用存储工厂获取并删除最新邮件
	if storageFactory == nil {
		c.JSON(500, gin.H{"error": "存储系统未初始化"})
		return
	}

	mail, err := storageFactory.GetStorage().GetAndDeleteLatestMail(mailHead)
	if err != nil {
		log.Printf("获取邮件失败: %v", err)
		c.JSON(500, gin.H{"error": "获取邮件失败"})
		return
	}

	if mail == nil {
		c.JSON(201, gin.H{"mail": "没有邮件"})
		return
	}

	c.JSON(200, gin.H{
		"mail": gin.H{
			"from":        mail.From,
			"title":       mail.Title,
			"TextContent": mail.TextContent,
			"HtmlContent": mail.HtmlContent,
		},
	})
}

func main() {
	// 初始化配置
	config = initConfig()

	// 设置日志格式
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	// 初始化存储工厂
	var err error
	storageFactory, err = NewStorageFactory()
	if err != nil {
		log.Fatalf("初始化存储工厂失败: %v", err)
	}

	log.Printf("存储模式: %s", storageFactory.GetMode())

	// 启动存储清理任务（仅对内存存储有效）
	storageFactory.StartCleanupTask()

	// 启动 HTTP 服务器
	go startHTTPServer()

	// 启动 SMTP 服务器
	if err := startSMTPServer(); err != nil {
		log.Fatalf("SMTP服务器启动失败: %v", err)
	}
}
