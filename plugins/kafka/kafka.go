package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"

	mesosproto "github.com/AVENTER-UG/mesos-m3s/proto"
	"github.com/AVENTER-UG/mesos-m3s/redis"
	"github.com/AVENTER-UG/util/util"
	"github.com/sirupsen/logrus"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Plugins struct {
	Redis    *redis.Redis
	Producer *kafka.Producer
	Topic    string
}

var plugin *Plugins

func Init(r *redis.Redis) string {
	plugin = &Plugins{
		Redis:    r,
		Producer: kafkaInit(),
		Topic:    util.Getenv("KAFKA_PLUGIN_TOPIC", "m3s_log"),
	}

	go plugin.kafkaEvent()

	logrus.WithField("func", "plugin.kafka.Event").Info("Kafka Plugin")

	return "kafka"
}

func Event(event mesosproto.Event) {
	logrus.WithField("func", "plugin.kafka.Event").Trace("Kafka Plugin")

	if plugin != nil {
		msg, err := json.Marshal(&event)
		if err != nil {
			logrus.WithField("func", "plugin.kafka.Event").Errorf("Could not marshal Mesos event: %s", err)
			return
		}

		topic := plugin.Topic + "_mesos"
		plugin.kafkaWrite("Event_Update", string(msg), &topic)
	}
}

func Logging(info string, args ...interface{}) {
	logrus.WithField("func", "plugin.kafka.Logging").Debug("Kafka Logging")
	if plugin != nil {
		msg, err := json.Marshal(&args)
		if err != nil {
			logrus.WithField("func", "plugin.kafka.Logging").Errorf("Could not marshal interface: %s", err)
			return
		}

		topic := plugin.Topic + "_logging"
		plugin.kafkaWrite(info, string(msg), &topic)
	}
}

func kafkaInit() *kafka.Producer {
	configFile := "kafka.properties"
	conf := ReadConfig(configFile)

	p, err := kafka.NewProducer(&conf)

	if err != nil {
		logrus.WithField("func", "plugin.kafka.kafka").Errorf("Failed to create producer: %s", err)
		return nil
	}

	return p
}

func (p *Plugins) kafkaEvent() {
	logrus.WithField("func", "plugin.kafka.kafkaEvent").Debug("Kafka Write Event")
	if p.Producer != nil {
		for e := range p.Producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					logrus.WithField("func", "plugin.kafka.kafkaEvent").Errorf("Failed to deliver message: %v", ev.TopicPartition)
				} else {
					logrus.WithField("func", "plugin.kafka.kafkaEvent").Debugf("Produced event to topic %s: key = %-10s value = %s",
						*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
				}
			}
		}
	}
}

func (p *Plugins) kafkaWrite(key string, data string, topic *string) {
	if p.Producer != nil {
		p.Producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: topic, Partition: kafka.PartitionAny},
			Key:            []byte(key),
			Value:          []byte(data),
		}, nil)

		p.Producer.Flush(15 * 1000)
	}
}

func ReadConfig(configFile string) kafka.ConfigMap {

	m := make(map[string]kafka.ConfigValue)

	file, err := os.Open(configFile)
	if err != nil {
		logrus.WithField("func", "plugin.kafka.ReadConfig").Errorf("Failed to open config file: %s", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") && len(line) != 0 {
			before, after, found := strings.Cut(line, "=")
			if found {
				parameter := strings.TrimSpace(before)
				value := strings.TrimSpace(after)
				m[parameter] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logrus.WithField("func", "plugin.kafka.ReadConfig").Errorf("Failed to read config file: %s", err)
		return nil
	}

	return m
}
