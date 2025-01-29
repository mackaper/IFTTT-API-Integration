package db

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type TeamPayload struct {
	Team int `json:"team"`
}

type triggerType interface {
	TeamPayload
}

var Ctx = context.Background()

type DBClient struct {
	Client *redis.Client
}

func MakeDBClient(address string, password string, number int) *DBClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       number,
	})
	dbclient := &DBClient{
		Client: rdb,
	}
	return dbclient
}

func keyifyPayload(payload TeamPayload) string {
	return fmt.Sprintf("team:%v", payload.Team)
}

func (db *DBClient) AddData(key string, data TeamPayload) {
	err := db.client.Set(ctx, keyifyPayload(data), key, 0).Err()
	if err != nil {
		panic(err)
	}
}

func (db *DBClient) GetData(data TeamPayload) string {
	val := db.client.Get(ctx, keyifyPayload(data)).Val()
	return val
}

func (db *DBClient) DeleteData(key string) {
	err := db.Client.Del(Ctx, key).Err()
	if err != nil {
		panic(err)
	}
}
