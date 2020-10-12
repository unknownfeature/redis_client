package cache

import (
	"context"
	"github.com/go-redis/redis"
	"log"
	"strings"
	"time"
)

func NewClient(db int, url string) *redis.Client {
	log.Println("connecting to the files", url, "db", db)
	cl := redis.NewClient(&redis.Options{
		Addr: url,
		DB:   db,
	})
	log.Println("connected to files")
	return cl
}

var get = func(client *redis.Client, ctx context.Context, params ...string) string {
	res := client.Get(params[0])
	return res.String()
}
var set = func(client *redis.Client, ctx context.Context, params ...string) string {
	var duration time.Duration
	if len(params) == 3 {
		converted, err := time.ParseDuration(params[2])
		if err != nil {
			log.Println("ignoring string parsing, default duration will be 0", err)
		} else {
			duration = converted
		}
	}
	res := client.Set(params[0], params[1], duration)
	return res.String()
}
var del = func(client *redis.Client, ctx context.Context, params ...string) string {
	res := client.Del(params...)
	return res.String()
}

var mget = func(client *redis.Client, ctx context.Context, params ...string) string {
	res := client.MGet(params...)
	if res.Err() == nil && res.Val() != nil && len(res.Val()) > 0{
		var sb strings.Builder
		for _, str := range res.Val() {
			sb.WriteString(str.(string))
			sb.WriteString("<br/>")
		}
		return sb.String()
	}
	return res.String()
}
var keys = func(client *redis.Client, ctx context.Context, params ...string) string {
	res := client.Keys(params[0])
	if res.Err() == nil && res.Val() != nil && len(res.Val()) > 0{
		var sb strings.Builder
		for _, str := range res.Val() {
			sb.WriteString(str)
			sb.WriteString("<br/>")
		}
		return sb.String()
	}
	return res.String()
}

const (
	Get  = "get"
	Set  = "set"
	Del  = "del"
	Mget = "mget"
	Keys = "keys"
)

var Commands = make(map[string]func(client *redis.Client, ctx context.Context, params ...string) string)

func init() {
	Commands[Get] = get
	Commands[Set] = set
	Commands[Del] = del
	Commands[Mget] = mget
	Commands[Keys] = keys
}
