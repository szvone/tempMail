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
	ms.mu.RLock()
	mails, exists := ms.mailBox[to]
	if !exists || len(mails) == 0 {
		ms.mu.RUnlock()
		return nil, nil
	}

	lastIndex := len(mails) - 1
	tmpMail := mails[lastIndex]
	ms.mu.RUnlock()

	ms.mu.Lock()
	// 重新检查最新状态，防止并发问题
	mails, exists = ms.mailBox[to]
	if exists && len(mails) > 0 {
		lastIndex := len(mails) - 1
		tmpMail = mails[lastIndex]
		if lastIndex == 0 {
			delete(ms.mailBox, to)
		} else {
			ms.mailBox[to] = mails[:lastIndex]
		}
	}
	ms.mu.Unlock()

	return &tmpMail, nil
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

	currentCount := ms.GetTotalMailCount()
	if currentCount <= limit {
		return 0, nil
	}

	toRemove := currentCount - limit + 100

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
		if mails, exists := ms.mailBox[key]; exists {
			if len(mails) > 0 {
				// 移除最早的邮件（索引0）
				ms.mailBox[key] = mails[1:]
				removed++
				if len(ms.mailBox[key]) == 0 {
					delete(ms.mailBox, key)
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
