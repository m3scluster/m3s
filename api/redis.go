package api

import (
	"context"
	"encoding/json"

	mesosutil "github.com/AVENTER-UG/mesos-util"
	mesosproto "github.com/AVENTER-UG/mesos-util/proto"
	goredis "github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// Redis struct about the redis connection
type Redis struct {
	RedisClient *goredis.Client
	RedisCTX    context.Context
}

// GetAllRedisKeys get out all redis keys to a patter
func (e *API) GetAllRedisKeys(pattern string) *goredis.ScanIterator {
	val := e.Redis.RedisClient.Scan(e.Redis.RedisCTX, 0, pattern, 0).Iterator()
	if err := val.Err(); err != nil {
		logrus.Error("getAllRedisKeys: ", err)
	}

	return val
}

// GetRedisKey get out the data of a key
func (e *API) GetRedisKey(key string) string {
	val, err := e.Redis.RedisClient.Get(e.Redis.RedisCTX, key).Result()
	if err != nil {
		logrus.Error("ge.Redis.RedisKey: ", err)
	}

	return val
}

// SetRedisKey store data in redis
func (e *API) SetRedisKey(data []byte, key string) {
	err := e.Redis.RedisClient.Set(e.Redis.RedisCTX, key, data, 0).Err()
	if err != nil {
		logrus.WithField("func", "SetRedisKey").Error("Could not save data in Redis: ", err.Error())
	}
}

// DelRedisKey will delete a redis key
func (e *API) DelRedisKey(key string) int64 {
	val, err := e.Redis.RedisClient.Del(e.Redis.RedisCTX, key).Result()
	if err != nil {
		logrus.Error("de.Redis.RedisKey: ", err)
		e.PingRedis()
	}

	return val
}

// GetTaskFromEvent get out the key by a mesos event
func (e *API) GetTaskFromEvent(update *mesosproto.Event_Update) mesosutil.Command {
	// search matched taskid in redis and update the status
	keys := e.GetAllRedisKeys(e.Framework.FrameworkName + ":*")
	for keys.Next(e.Redis.RedisCTX) {
		// ignore redis keys if they are not mesos tasks
		if keys.Val() == e.Framework.FrameworkName+":framework" || keys.Val() == e.Framework.FrameworkName+":framework_config" {
			continue
		}
		// get the values of the current key
		key := e.GetRedisKey(keys.Val())

		task := mesosutil.DecodeTask(key)

		if task.TaskID == update.Status.TaskID.Value {
			task.State = update.Status.State.String()
			return task
		}
	}

	return mesosutil.Command{}
}

// CountRedisKey will get back the count of the redis key
func (e *API) CountRedisKey(pattern string) int {
	keys := e.GetAllRedisKeys(pattern)
	count := 0
	for keys.Next(e.Redis.RedisCTX) {
		count++
	}
	logrus.Debug("CountRedisKey: ", pattern, count)
	return count
}

// SaveConfig store the current framework config
func (e *API) SaveConfig() {
	data, _ := json.Marshal(e.Config)
	err := e.Redis.RedisClient.Set(e.Redis.RedisCTX, e.Framework.FrameworkName+":framework_config", data, 0).Err()
	if err != nil {
		logrus.Error("Framework save config state into redis error:", err)
	}
}

// PingRedis to check the health of redis
func (e *API) PingRedis() error {
	pong, err := e.Redis.RedisClient.Ping(e.Redis.RedisCTX).Result()
	if err != nil {
		logrus.WithField("func", "PingRedis").Error("Did not pon Redis: ", pong, err.Error())
	}
	if err != nil {
		return err
	}
	return nil
}

// ConnectRedis will connect the redis DB and save the client pointer
func (e *API) ConnectRedis() {
	var redisOptions goredis.Options
	redisOptions.Addr = e.Config.RedisServer
	redisOptions.DB = e.Config.RedisDB
	if e.Config.RedisPassword != "" {
		redisOptions.Password = e.Config.RedisPassword
	}

	e.Redis.RedisClient = goredis.NewClient(&redisOptions)
	e.Redis.RedisCTX = context.Background()

	err := e.PingRedis()
	if err != nil {
		e.Redis.RedisClient.Close()
		e.ConnectRedis()
	}
}

// SaveTaskRedis store mesos task in DB
func (e *API) SaveTaskRedis(cmd mesosutil.Command) {
	d, _ := json.Marshal(&cmd)
	e.SetRedisKey(d, cmd.TaskName+":"+cmd.TaskID)
}

// SaveFrameworkRedis store mesos framework in DB
func (e *API) SaveFrameworkRedis() {
	d, _ := json.Marshal(&e.Framework)
	e.SetRedisKey(d, e.Framework.FrameworkName+":framework")
}
