package data

import (
	"errors"
	"log"

	"github.com/mdlayher/goat/goat/common"

	"github.com/garyburd/redigo/redis"
)

// RedisConnect initiates a connection to Redis server
func RedisConnect() (c redis.Conn, err error) {
	c, err = redis.Dial("tcp", common.Static.Config.Redis.Host)
	if err != nil {
		return
	}

	// Authenticate with Redis database if necessary
	if common.Static.Config.Redis.Password != "" {
		_, err = c.Do("AUTH", common.Static.Config.Redis.Password)
	}
	return
}

// RedisPing verifies that Redis server is available
func RedisPing() bool {
	// Send redis a PING request
	reply, err := RedisDo("PING")
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// Ensure value is valid
	res, err := redis.String(reply, nil)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// PONG is valid response
	if res != "PONG" {
		log.Println("redisPing: redis replied to PING with:", res)
		return false
	}

	// Redis OK
	return true
}

// RedisDo runs a single Redis command and returns its reply
func RedisDo(command string, args ...interface{}) (interface{}, error) {
	// Open Redis connection
	c, err := RedisConnect()
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("redisDo: failed to connect to redis")
	}

	// Send Redis command with arguments, receive reply
	reply, err := c.Do(command, args...)
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("redisDo: failed to send command to redis: " + command)
	}

	if err := c.Close(); err != nil {
		log.Println(err.Error())
	}

	return reply, nil
}
