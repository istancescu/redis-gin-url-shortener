package config

import (
	"bufio"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type RedisConfig struct {
	Server struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
}

func CreateNewRedisConfig(path string) *redis.Options {
	redisConfig := new(RedisConfig)

	config, err := os.Open(path)

	if err != nil {
		log.Panicf("Couldn't find file for path: %s \n", path)
	}

	defer func(config *os.File) {
		err := config.Close()
		if err != nil {
			log.Panicln("Couldn't close file!")
		}
	}(config)

	reader := bufio.NewReader(config)

	decoder := yaml.NewDecoder(reader)

	err = decoder.Decode(&redisConfig)
	if err != nil {
		log.Panicln("There was a problem decoding the config file!")
	}

	return &redis.Options{
		Addr:     redisConfig.Server.Host + ":" + redisConfig.Server.Port,
		Password: redisConfig.Server.Password,
		DB:       redisConfig.Server.DB,
	}
}
