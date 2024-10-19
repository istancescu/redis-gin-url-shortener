package config

import (
	"bufio"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
	"os"
)

type RedisConfig struct {
	Redis struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
}

func ProvideRedisConfig(path string) (*redis.Options, error) {
	redisConfig, err := loadConfigFromAFile(path)

	if err != nil {
		return nil, err
	}

	return &redis.Options{
		Addr:     redisConfig.Redis.Host + ":" + redisConfig.Redis.Port,
		Password: redisConfig.Redis.Password,
		DB:       redisConfig.Redis.DB,
	}, nil
}

func loadConfigFromAFile(path string) (*RedisConfig, error) {
	redisConfig := new(RedisConfig)

	config, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("couldn't find file for path: %s \n", path)
	}

	defer config.Close()

	reader := bufio.NewReader(config)
	decoder := yaml.NewDecoder(reader)

	err = decoder.Decode(&redisConfig)

	if err != nil {
		return nil, fmt.Errorf("there was a problem decoding the config file")
	}
	return redisConfig, nil
}
