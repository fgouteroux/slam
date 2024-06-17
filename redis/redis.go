package redis

import (
	"fmt"
	"time"

	redigo "github.com/gomodule/redigo/redis"
)

func Connect(host string, db int) *redigo.Pool {
	return &redigo.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,

		Dial: func() (redigo.Conn, error) {
			c, err := redigo.Dial("tcp", host, redigo.DialDatabase(db))
			if err != nil {
				return nil, err
			}
			return c, err
		},

		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func Ping(db *redigo.Pool) error {
	conn := db.Get()
	defer conn.Close()

	_, err := redigo.String(conn.Do("PING"))
	if err != nil {
		return fmt.Errorf("Failed to ping: %v", err)
	}
	return nil
}

func Get(db *redigo.Pool, key string) ([]byte, error) {
	conn := db.Get()
	defer conn.Close()

	var data []byte
	data, err := redigo.Bytes(conn.Do("GET", PrefixKey(key)))
	if err != nil {
		return data, fmt.Errorf("Error getting key %s: %v", key, err)
	}
	return data, err
}

func Set(db *redigo.Pool, key string, value []byte, ttl int) error {
	conn := db.Get()
	defer conn.Close()

	var err error

	if ttl > 0 {
		_, err = conn.Do("SET", PrefixKey(key), value, "EX", ttl)
	} else {
		_, err = conn.Do("SET", PrefixKey(key), value)
	}

	if err != nil {
		return fmt.Errorf("Failed to set key %s: %v", key, err)
	}
	return err
}

func Exists(db *redigo.Pool, key string) (bool, error) {
	conn := db.Get()
	defer conn.Close()

	ok, err := redigo.Bool(conn.Do("EXISTS", PrefixKey(key)))
	if err != nil {
		return ok, fmt.Errorf("Failed to check if key %s exists: %v", key, err)
	}
	return ok, err
}

func List(db *redigo.Pool, key string) ([]interface{}, error) {
	conn := db.Get()
	defer conn.Close()

	data, err := redigo.Values(conn.Do("KEYS", PrefixKey(key)))
	if err != nil {
		return data, fmt.Errorf("Failed to list key %s: %v", key, err)
	}

	return data, err
}

func Delete(db *redigo.Pool, key string) (int, error) {
	conn := db.Get()
	defer conn.Close()

	val, err := redigo.Int(conn.Do("DEL", PrefixKey(key)))
	if err != nil {
		return val, fmt.Errorf("Failed to delete key %s: %v", key, err)
	}
	return val, err
}

func PrefixKey(key string) string {
	return "slam:" + key
}
