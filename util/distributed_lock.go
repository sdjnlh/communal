package util

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
)
var pool = &redis.Pool{
	Wait:true,
	Dial: func() (conn redis.Conn, e error) {
		conn, e = redis.Dial("tcp", "127.0.0.1:6379")
		if e != nil {
			return nil, e
		}
		return
	},
}

func getConn() redis.Conn{
	return pool.Get()
}

func key(str string) string{
	return fmt.Sprintf("distributed_lock:%s", str)
}

func TryLock(str, token string, expire int64) (ok bool, err error){
	conn := getConn()
	defer conn.Close()
	_, err = redis.String(conn.Do("SET", key(str), token, "EX", expire, "NX"))
	if err == redis.ErrNil{
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func Unlock(str string) (err error){
	conn := getConn()
	defer conn.Close()
	_, err = conn.Do("DEL", key(str))
	return
}

func AddTimeout(str, token string, expire int64) (err error) {
	conn := getConn()
	defer conn.Close()
	ttl, err := redis.Int64(conn.Do("TTL", key(str)))
	if err != nil{
		log.Fatal("ttl failed ,error:", err)
	}
	if ttl > 0 {
		_, err := redis.String(conn.Do("SET", key(str), token, "EX", int(ttl+expire)))
		if err == redis.ErrNil {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}