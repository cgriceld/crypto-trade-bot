package main

import (
	"fmt"
	"os"
	"strconv"
)

type config struct {
	port       string
	dsn        string
	APIPublic  string
	APIPrivate string
	TgBotURL   string
	TgChatID   int
}

func configApp() (*config, error) {
	c := &config{}

	getConf := func(key string) (string, error) {
		val, _ := os.LookupEnv(key)
		if val == "" {
			return val, fmt.Errorf("No config: %s", key)
		} else {
			return val, nil
		}
	}

	if val, err := getConf("APIPublic"); err == nil {
		c.APIPublic = val
	} else {
		return nil, err
	}

	if val, err := getConf("APIPrivate"); err == nil {
		c.APIPrivate = val
	} else {
		return nil, err
	}

	if val, err := getConf("dsn"); err == nil {
		c.dsn = val
	} else {
		return nil, err
	}

	if val, err := getConf("port"); err == nil {
		c.port = val
	} else {
		return nil, err
	}

	if val, err := getConf("TgBotURL"); err == nil {
		c.TgBotURL = val
	} else {
		return nil, err
	}

	if val, err := getConf("TgChatID"); err == nil {
		if id, err := strconv.Atoi(val); err == nil {
			c.TgChatID = id
		} else {
			return nil, fmt.Errorf("Fail to convert TgChatID")
		}
	} else {
		return nil, err
	}

	return c, nil
}
