package redis

import (
	"context"

	cfg "github.com/AVENTER-UG/mesos-m3s/controller/types"
	goredis "github.com/go-redis/redis/v9"

	"github.com/sirupsen/logrus"
)

// Redis struct about the redis connection
type Redis struct {
	Client   *goredis.Client
	CTX      context.Context
	Server   string
	Password string
	DB       int
	Prefix   string
}

// New will create a new Redis object
func New(cfg *cfg.Config) *Redis {
	e := &Redis{
		Server:   cfg.RedisServer,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
		Prefix:   cfg.RedisPrefix,
		CTX:      context.Background(),
	}

	logrus.WithField("func", "Redis.New").Info("Redis Connection", e.Connect())

	return e
}

// GetAllRedisKeys get out all redis keys to a patter
func (e *Redis) GetAllRedisKeys(pattern string) *goredis.ScanIterator {
	val := e.Client.Scan(e.CTX, 0, pattern, 0).Iterator()
	if err := val.Err(); err != nil {
		logrus.Error("getAllRedisKeys: ", err)
	}

	return val
}

// GetRedisKey get out the data of a key
func (e *Redis) GetRedisKey(key string) string {
	val, _ := e.Client.Get(e.CTX, key).Result()
	return val
}

// SetRedisKey store data in redis
func (e *Redis) SetRedisKey(data []byte, key string) {
	err := e.Client.Set(e.CTX, key, data, 0).Err()
	if err != nil {
		logrus.WithField("func", "SetRedisKey").Error("Could not save data in Redis: ", err.Error())
	}
}

// DelRedisKey will delete a redis key
func (e *Redis) DelRedisKey(key string) int64 {
	val, err := e.Client.Del(e.CTX, key).Result()
	if err != nil {
		logrus.Error("de.Key: ", err)
		e.PingRedis()
	}

	return val
}

// PingRedis to check the health of redis
func (e *Redis) PingRedis() error {
	pong, err := e.Client.Ping(e.CTX).Result()
	if err != nil {
		logrus.WithField("func", "PingRedis").Error("Did not pon Redis: ", pong, err.Error())
	}
	if err != nil {
		return err
	}
	return nil
}

// Connect will connect the redis DB and save the client pointer
func (e *Redis) Connect() bool {
	var redisOptions goredis.Options
	redisOptions.Addr = e.Server
	redisOptions.DB = e.DB
	if e.Password != "" {
		redisOptions.Password = e.Password
	}

	e.Client = goredis.NewClient(&redisOptions)

	err := e.PingRedis()
	if err != nil {
		e.Client.Close()
	}

	return true
}
