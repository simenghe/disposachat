package main

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

func Ping(c redis.Conn) error {
	s, err := redis.String(c.Do("PING"))
	fmt.Println(s)
	return err
}

func Setex(c redis.Conn, key string, value string, timeMinutes int) (string, error) {
	s, err := redis.String(c.Do("SETEX", key, timeMinutes*60, value))
	return s, err
}

func Keys(c redis.Conn) ([]string, error) {
	s, err := redis.Strings(c.Do("KEYS", "*"))
	return s, err
}

func Get(c redis.Conn, key string) (string, error) {
	s, err := redis.String(c.Do("GET", key))
	return s, err
}
