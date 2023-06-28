package db

import (
	"github.com/gomodule/redigo/redis"
)

var Redis *redis.Pool

func SetKeyWithExpireTime(con *redis.Pool, key string, value string, time int) bool {
	c := con.Get()
	defer c.Close()
	_, err := c.Do("SETEX", key, time, value)
	if err != nil {
		return false
	}
	return true
}

func RedisGet(con *redis.Pool, key string) (string, error) {
	c := con.Get()
	defer c.Close()
	return redis.String(c.Do("GET", key))
}

func RedisGetAndRefreshExpireTime(con *redis.Pool, key string, time int) (string, error) {
	c := con.Get()
	defer c.Close()
	str, err := redis.String(c.Do("GET", key))
	if str != "" && err == nil {
		res, rerr := redis.Int(c.Do("EXPIRE", key, time))
		if res == 1 {
			return str, rerr
		}
		return "", rerr
	}
	return "", err
}
