package api

import (
	"context"
	"encoding/json"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	mesosproto "github.com/AVENTER-UG/mesos-util/proto"
	goredis "github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// GetAllRedisKeys get out all keys in redis depends to the pattern
func GetAllRedisKeys(pattern string) *goredis.ScanIterator {
	val := config.RedisClient.Scan(config.RedisCTX, 0, pattern, 0).Iterator()
	if err := val.Err(); err != nil {
		logrus.Error("getAllRedisKeys: ", err)
	}
	return val
}

// GetRedisKey get out all values to a key
func GetRedisKey(key string) string {
	val, err := config.RedisClient.Get(config.RedisCTX, key).Result()
	if err != nil {
		logrus.Error("getRedisKey: ", err)
	}
	return val
}

// DelRedisKey will delete a redis key
func DelRedisKey(key string) int64 {
	val, err := config.RedisClient.Del(config.RedisCTX, key).Result()
	if err != nil {
		logrus.Error("delRedisKey: ", err)
	}

	return val
}

// GetTaskFromEvent get out the task to an event
func GetTaskFromEvent(update *mesosproto.Event_Update) mesosutil.Command {
	// search matched taskid in redis and update the status
	keys := GetAllRedisKeys("*")
	for keys.Next(config.RedisCTX) {
		// get the values of the current key
		key := GetRedisKey(keys.Val())

		// update the status of the matches task
		var task mesosutil.Command
		json.Unmarshal([]byte(key), &task)
		if task.TaskID == update.Status.TaskID.Value {
			task.State = update.Status.State.String()

			return task
		}
	}
	return mesosutil.Command{}
}

// CountRedisKey will get back the count of the redis key
func CountRedisKey(pattern string) int {
	keys := GetAllRedisKeys(pattern)
	count := 0
	for keys.Next(config.RedisCTX) {
		count++
	}
	logrus.Debug("CountRedisKey: ", pattern, count)
	return count
}

// SaveConfig store the current framework config
func SaveConfig() {
	data, _ := json.Marshal(config)
	err := config.RedisClient.Set(config.RedisCTX, framework.FrameworkName+":framework_config", data, 0).Err()
	if err != nil {
		logrus.Error("Framework save config and state into redis Error: ", err)
	}
}

// PingRedis to check the health of redis
func PingRedis() error {
	pong, err := config.RedisClient.Ping(config.RedisCTX).Result()
	logrus.Debug("Redis Health: ", pong, err)
	if err != nil {
		return err
	}
	return nil
}

// ConnectRedis will connect the redis DB and save the client pointer
func ConnectRedis() {
	var redisOptions goredis.Options
	redisOptions.Addr = config.RedisServer
	redisOptions.DB = config.RedisDB
	if config.RedisPassword != "" {
		redisOptions.Password = config.RedisPassword
	}

	config.RedisClient = goredis.NewClient(&redisOptions)
	config.RedisCTX = context.Background()

	err := PingRedis()
	if err != nil {
		ConnectRedis()
	}
}
