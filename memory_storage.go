package main

import (
	"sync"
)

type MemoryStorage struct {
	mailBox map[string][]MailContent
	mu      sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		mailBox: make(map[string][]MailContent),
	}
}

func (ms *MemoryStorage) StoreMail(to string, content MailContent) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, ok := ms.mailBox[to]; !ok {
		ms.mailBox[to] = make([]MailContent, 0, 10)
	}
	ms.mailBox[to] = append(ms.mailBox[to], content)
	return nil
}

func (ms *MemoryStorage) GetAndDeleteLatestMail(to string) (*MailContent, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	mails, exists := ms.mailBox[to]
	if !exists || len(mails) == 0 {
		return nil, nil
	}

	// 获取最新的邮件（最后添加的）
	lastIndex := len(mails) - 1
	mail := mails[lastIndex]

	// 删除该邮件
	if lastIndex == 0 {
		delete(ms.mailBox, to)
	} else {
		ms.mailBox[to] = mails[:lastIndex]
	}

	return &mail, nil
}

func (ms *MemoryStorage) GetTotalMailCount() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	total := 0
	for _, mails := range ms.mailBox {
		total += len(mails)
	}
	return total
}

func (ms *MemoryStorage) CleanupOldestMails(limit int) (int, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	currentCount := 0
	for _, mails := range ms.mailBox {
		currentCount += len(mails)
	}

	if currentCount <= limit {
		return 0, nil
	}

	toRemove := currentCount - limit
	if toRemove <= 0 {
		return 0, nil
	}

	// 收集所有邮件并按时间排序
	type mailWithKey struct {
		key  string
		mail MailContent
	}

	allMails := make([]mailWithKey, 0)
	for key, mails := range ms.mailBox {
		for _, mail := range mails {
			allMails = append(allMails, mailWithKey{key: key, mail: mail})
		}
	}

	// 按接收时间升序排序（最早的在前）
	for i := 0; i < len(allMails)-1; i++ {
		for j := i + 1; j < len(allMails); j++ {
			if allMails[i].mail.ReceivedAt.After(allMails[j].mail.ReceivedAt) {
				allMails[i], allMails[j] = allMails[j], allMails[i]
			}
		}
	}

	removed := 0
	for i := 0; i < len(allMails) && removed < toRemove; i++ {
		key := allMails[i].key
		if mails, exists := ms.mailBox[key]; exists && len(mails) > 0 {
			// 找到要删除的邮件在列表中的位置
			for j := 0; j < len(mails); j++ {
				if mails[j].ReceivedAt.Equal(allMails[i].mail.ReceivedAt) {
					// 删除该邮件
					if len(mails) == 1 {
						delete(ms.mailBox, key)
					} else {
						ms.mailBox[key] = append(mails[:j], mails[j+1:]...)
					}
					removed++
					break
				}
			}
		}
	}

	return removed, nil
}

func (ms *MemoryStorage) ClearAllMails() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.mailBox = make(map[string][]MailContent)
	return nil
}

func (ms *MemoryStorage) HasMail(to string) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	mails, exists := ms.mailBox[to]
	return exists && len(mails) > 0
}
