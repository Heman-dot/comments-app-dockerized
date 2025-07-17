package redisc

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"log"
	"strconv"
)

var Client *redis.Client

func getIndex(ctx context.Context) (int, error) {
	val, err := Client.Get(ctx, "index").Result()
	var i int
	if err != nil && err != redis.Nil {
		log.Printf("Problem with redis: %s", err)
		return -1, err
	} else if err == redis.Nil {
		i = 0
	} else {
		i, err = strconv.Atoi(val)
		if err != nil {
			log.Printf("Problem in stored data: %s", err)
			return -1, err
		}
	}
	return i, nil
}

func getIndexAndBump(ctx context.Context) (int, error) {
	i, err := getIndex(ctx)
	if err != nil {
		return -1, err
	}
	err = Client.Set(ctx, "index", strconv.Itoa(i+1), 0).Err()
	if err != nil {
		log.Printf("Problem with redis while trying to store data: %s", err)
		return -1, err
	}
	return i + 1, err
}

func StoreComment(comment string) (int, error) {
	ctx := context.Background()
	i, err := getIndexAndBump(ctx)
	if err != nil {
		return -1, err
	}
	err = Client.Set(ctx, "comment"+strconv.Itoa(i), comment, 0).Err()
	if err != nil {
		log.Printf("Problem with redis while trying to store data: %s", err)
		return -1, err
	}
	return i, err
}

func GetLatestComments() (string, error) {
	messages := make([]string, 0)
	ctx := context.Background()
	index, err := getIndex(ctx)
	if err != nil {
		return "", err
	}
	for i := 0; i < 10; i++ {
		if index-i <= 0 {
			break
		}
		val, err := Client.Get(ctx, "comment"+strconv.Itoa(index-i)).Result()
		if err != nil && err != redis.Nil {
			log.Printf("Problem with redis while trying to retrieve data: %s", err)
			return "", err
		} else if err == redis.Nil {
			continue
		}
		messages = append(messages, val)
	}
	msg, err := json.Marshal(messages)
	if err != nil {
		log.Fatalf("Json issue")
	}
	return string(msg), err
}
