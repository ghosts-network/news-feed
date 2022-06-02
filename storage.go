package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
)

type NewsStorage interface {
	AddPublication(publication *Publication)
	// UpdatePublication(publication Publication)
	RemovePublication(publication *Publication)

	//GetUserSources()
	//AddUserSource()
	//RemoveUserSource()

	FindNews(user string, cursor string) []Publication
}

type RedisNewsStorage struct {
	rdb *redis.Client
}

func (storage *RedisNewsStorage) AddPublication(publication *Publication) {
	storage.rdb.HSetNX(context.Background(), "user:1:news", publication.Id, publication.Content).Result()
}

func (storage *RedisNewsStorage) RemovePublication(publication *Publication) {
	storage.rdb.HDel(context.Background(), "user:1:news", publication.Id).Result()
}

func (storage *RedisNewsStorage) FindNews(user string, cursor string) []Publication {
	key := fmt.Sprintf("user:%v:news", user)
	resp, err := storage.rdb.HGetAll(context.Background(), key).Result()
	if err != nil {
		log.Println(resp)
	}

	return nil
}

func NewRedisNewsStorage(rdb *redis.Client) *RedisNewsStorage {
	return &RedisNewsStorage{rdb: rdb}
}
