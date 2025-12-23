package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// StorageFactory 存储工厂，负责创建和管理存储实例
type StorageFactory struct {
	storage Storage
	mode    string
}

// NewStorageFactory 创建存储工厂实例
func NewStorageFactory() (*StorageFactory, error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("STORAGE_MODE")))
	if mode == "" {
		mode = "memory" // 默认使用内存模式
	}

	factory := &StorageFactory{mode: mode}
	
	switch mode {
	case "memory":
		factory.storage = NewMemoryStorage()
		log.Printf("使用内存存储模式")
		
	case "redis":
		addr := os.Getenv("REDIS_ADDR")
		if addr == "" {
			addr = "localhost:6379"
		}
		password := os.Getenv("REDIS_PASSWORD")
		dbStr := os.Getenv("REDIS_DB")
		db := 0
		if dbStr != "" {
			if parsedDB, err := strconv.Atoi(dbStr); err == nil {
				db = parsedDB
			}
		}
		
		factory.storage = NewRedisStorage(addr, password, db)
		log.Printf("使用Redis存储模式，地址: %s, 数据库: %d", addr, db)
		
	case "mysql":
		dsn := os.Getenv("MYSQL_DSN")
		if dsn == "" {
			dsn = "root:password@tcp(localhost:3306)/temp_mail?charset=utf8mb4&parseTime=True&loc=Local"
		}
		
		mysqlStorage, err := NewMySQLStorage(dsn)
		if err != nil {
			return nil, fmt.Errorf("创建MySQL存储失败: %v", err)
		}
		factory.storage = mysqlStorage
		log.Printf("使用MySQL存储模式，DSN: %s", dsn)
		
	default:
		return nil, fmt.Errorf("不支持的存储模式: %s，支持的存储模式: memory, redis, mysql", mode)
	}
	
	return factory, nil
}

// GetStorage 获取当前存储实例
func (sf *StorageFactory) GetStorage() Storage {
	return sf.storage
}

// GetMode 获取当前存储模式
func (sf *StorageFactory) GetMode() string {
	return sf.mode
}

// StartCleanupTask 启动清理任务（仅对内存存储有效）
func (sf *StorageFactory) StartCleanupTask() {
	if sf.mode != "memory" {
		log.Printf("存储模式为 %s，使用自动过期/软删除机制，不需要定时清理", sf.mode)
		return
	}
	
	intervalStr := os.Getenv("MEMORY_CLEANUP_INTERVAL")
	interval := 5 // 默认5分钟
	if intervalStr != "" {
		if parsedInterval, err := strconv.Atoi(intervalStr); err == nil && parsedInterval > 0 {
			interval = parsedInterval
		}
	}
	
	log.Printf("启动内存存储清理任务，间隔: %d 分钟", interval)
	
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Minute)
		defer ticker.Stop()
		
		// 立即执行一次清理
		sf.cleanupMemoryStorage()
		
		for {
			select {
			case <-ticker.C:
				sf.cleanupMemoryStorage()
			}
		}
	}()
}

// cleanupMemoryStorage 清理内存存储
func (sf *StorageFactory) cleanupMemoryStorage() {
	if sf.mode != "memory" {
		return
	}
	
	if memoryStorage, ok := sf.storage.(*MemoryStorage); ok {
		removed, err := memoryStorage.CleanupOldestMails(1000) // 清理到1000封邮件以下
		if err != nil {
			log.Printf("内存存储清理失败: %v", err)
		} else if removed > 0 {
			log.Printf("内存存储清理完成，移除了 %d 封邮件，当前总数: %d", removed, memoryStorage.GetTotalMailCount())
		}
	}
}