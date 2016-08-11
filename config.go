package main

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	EtcdEndpoints   []string
	ServiceDir      string `ymal:"ServiceDir"`
	CheckTimeout    int
	CheckInterval   int
	Concurrency     int
	MaxFails        int
	RetryDelay      int
	OkStatus        []int
	DefaultCheckUrl string
}

func ReadConfig(fpath string) (*Config, error) {
	c := Config{}

	if fpath == "" {
		return &c, errors.New("No config file provided.")
	}

	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		return &c, err
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return &c, err
	}

	return &c, nil
}
