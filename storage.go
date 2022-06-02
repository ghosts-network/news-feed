package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
)

type NewsStorage interface {
	AddPublication(publication *Publication) error
	UpdatePublication(publication *Publication) error
	RemovePublication(publication *Publication) error

	//GetUserSources()
	//AddUserSource()
	//RemoveUserSource()

	FindNews(user string, cursor string) []Publication
}

type RedisNewsStorage struct {
	rdb *redis.Client
}

func (storage *RedisNewsStorage) AddPublication(publication *Publication) (err error) {
	_, err = storage.rdb.HSet(context.Background(), "user:1:news", publication.Id, publication.Content).Result()
	return
}

func (storage *RedisNewsStorage) UpdatePublication(publication *Publication) (err error) {
	_, err = storage.rdb.HSet(context.Background(), "user:1:news", publication.Id, publication.Content).Result()
	return
}

func (storage *RedisNewsStorage) RemovePublication(publication *Publication) (err error) {
	_, err = storage.rdb.HDel(context.Background(), "user:1:news", publication.Id).Result()
	return
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
