package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisStorage(addr, password string, db int) *RedisStorage {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisStorage{
		client: client,
		ctx:    context.Background(),
	}
}

func (rs *RedisStorage) StoreMail(to string, content MailContent) error {
	// 使用Redis的List结构，每个邮箱对应一个List，存储序列化的MailContent
	key := fmt.Sprintf("mailbox:%s", to)
	data, err := json.Marshal(content)
	if err != nil {
		return err
	}

	// 按时间戳排序，将最新的邮件放在列表头部（左边）
	_, err = rs.client.LPush(rs.ctx, key, data).Result()
	if err != nil {
		return err
	}

	// 设置5分钟过期时间，自动失效
	_, err = rs.client.Expire(rs.ctx, key, 5*time.Minute).Result()
	return err
}

func (rs *RedisStorage) GetAndDeleteLatestMail(to string) (*MailContent, error) {
	key := fmt.Sprintf("mailbox:%s", to)

	// 从列表右侧弹出最新邮件（LPush + RPop = 最新邮件在左边）
	data, err := rs.client.RPop(rs.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			// 没有邮件
			return nil, nil
		}
		return nil, err
	}

	var mail MailContent
	err = json.Unmarshal(data, &mail)
	if err != nil {
		return nil, err
	}

	return &mail, nil
}

func (rs *RedisStorage) GetTotalMailCount() int {
	// 扫描所有mailbox:*键并统计总数
	keys, err := rs.client.Keys(rs.ctx, "mailbox:*").Result()
	if err != nil {
		log.Printf("Redis keys scan error: %v", err)
		return 0
	}

	total := 0
	for _, key := range keys {
		count, err := rs.client.LLen(rs.ctx, key).Result()
		if err != nil {
			log.Printf("Redis llen error: %v", err)
			continue
		}
		total += int(count)
	}

	return total
}

func (rs *RedisStorage) CleanupOldestMails(limit int) (int, error) {
	// Redis存储使用自动过期机制，不需要手动清理
	log.Printf("Redis存储使用自动过期机制，跳过手动清理")
	return 0, nil
}

func (rs *RedisStorage) ClearAllMails() error {
	keys, err := rs.client.Keys(rs.ctx, "mailbox:*").Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		_, err = rs.client.Del(rs.ctx, keys...).Result()
		if err != nil {
			return err
		}
	}
	return nil
}

func (rs *RedisStorage) HasMail(to string) bool {
	key := fmt.Sprintf("mailbox:%s", to)
	count, err := rs.client.LLen(rs.ctx, key).Result()
	if err != nil {
		return false
	}
	return count > 0
}
