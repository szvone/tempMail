package main

import "time"

type MailContent struct {
	From        string    `json:"from"`
	To          string    `json:"to"`
	Title       string    `json:"title"`
	TextContent string    `json:"TextContent"`
	HtmlContent string    `json:"HtmlContent"`
	ReceivedAt  time.Time `json:"receivedAt"`
}

type Storage interface {
	// 存储邮件
	StoreMail(to string, content MailContent) error
	// 获取并移除最新邮件
	GetAndDeleteLatestMail(to string) (*MailContent, error)
	// 获取所有邮件总数
	GetTotalMailCount() int
	// 清理最早的邮件，直到总数降到limit以下
	CleanupOldestMails(limit int) (int, error)
	// 清空所有邮件
	ClearAllMails() error
	// 检查是否存在邮件
	HasMail(to string) bool
}
