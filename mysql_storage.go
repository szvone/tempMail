package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLStorage struct {
	db *sql.DB
}

func NewMySQLStorage(dsn string) (*MySQLStorage, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// 初始化数据库表
	if err := initMySQLTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return &MySQLStorage{db: db}, nil
}

func initMySQLTables(db *sql.DB) error {
	// 创建邮件表
	query := `
	CREATE TABLE IF NOT EXISTS mails (
		id INT AUTO_INCREMENT PRIMARY KEY,
		to_email VARCHAR(255) NOT NULL,
		from_email VARCHAR(255) NOT NULL,
		title VARCHAR(255) NOT NULL,
		text_content TEXT,
		html_content TEXT,
		received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_read TINYINT(1) DEFAULT 0 COMMENT '是否已读取',
		read_at TIMESTAMP NULL COMMENT '读取时间',
		INDEX idx_to_email (to_email),
		INDEX idx_received_at (received_at),
		INDEX idx_is_read (is_read)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	_, err := db.Exec(query)
	return err
}

func (ms *MySQLStorage) StoreMail(to string, content MailContent) error {
	query := `INSERT INTO mails (to_email, from_email, title, text_content, html_content, received_at, is_read) VALUES (?, ?, ?, ?, ?, ?, 0)`
	_, err := ms.db.Exec(query, to, content.From, content.Title, content.TextContent, content.HtmlContent, content.ReceivedAt)
	return err
}

func (ms *MySQLStorage) GetAndDeleteLatestMail(to string) (*MailContent, error) {
	// 使用事务确保原子性：先查询最新未读邮件，然后标记为已读
	tx, err := ms.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 查询最新未读邮件
	query := `SELECT id, from_email, title, text_content, html_content, received_at FROM mails WHERE to_email = ? AND is_read = 0 ORDER BY received_at DESC LIMIT 1`
	row := tx.QueryRow(query, to)

	var id int
	var mail MailContent
	err = row.Scan(&id, &mail.From, &mail.Title, &mail.TextContent, &mail.HtmlContent, &mail.ReceivedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Rollback()
			return nil, nil
		}
		tx.Rollback()
		return nil, err
	}

	// 标记邮件为已读（软删除）
	updateQuery := `UPDATE mails SET is_read = 1, read_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err = tx.Exec(updateQuery, id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return &mail, nil
}

func (ms *MySQLStorage) GetTotalMailCount() int {
	var count int
	query := `SELECT COUNT(*) FROM mails WHERE is_read = 0`
	err := ms.db.QueryRow(query).Scan(&count)
	if err != nil {
		log.Printf("MySQL count error: %v", err)
		return 0
	}
	return count
}

func (ms *MySQLStorage) CleanupOldestMails(limit int) (int, error) {
	// MySQL存储使用软删除机制，不需要手动清理
	log.Printf("MySQL存储使用软删除机制，跳过手动清理")
	return 0, nil
}

func (ms *MySQLStorage) ClearAllMails() error {
	query := `UPDATE mails SET is_read = 1, read_at = CURRENT_TIMESTAMP WHERE is_read = 0`
	_, err := ms.db.Exec(query)
	return err
}

func (ms *MySQLStorage) HasMail(to string) bool {
	var count int
	query := `SELECT COUNT(*) FROM mails WHERE to_email = ? AND is_read = 0`
	err := ms.db.QueryRow(query, to).Scan(&count)
	if err != nil {
		log.Printf("MySQL has mail error: %v", err)
		return false
	}
	return count > 0
}
