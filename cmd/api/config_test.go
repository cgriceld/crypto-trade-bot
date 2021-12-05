package main

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Conf struct {
	name string
	res  *config
	err  error
	set  map[string]string
}

var (
	full = &config{
		port:       "123",
		dsn:        "123",
		APIPublic:  "123",
		APIPrivate: "123",
		TgBotURL:   "123",
		TgChatID:   123,
	}
)

func TestConfig(t *testing.T) {
	tests := []Conf{
		{"All Set", full, nil, map[string]string{}},
		{"All Set", nil, errors.New("No config: port"), map[string]string{"port": ""}},
		{"All Set", nil, errors.New("Fail to convert TgChatID"), map[string]string{"TgChatID": "fff"}},
	}

	os.Setenv("dsn", "123")
	os.Setenv("APIPublic", "123")
	os.Setenv("APIPrivate", "123")
	os.Setenv("TgBotURL", "123")
	os.Setenv("TgChatID", "123")

	for _, test := range tests {
		os.Setenv("port", "123")
		for k, v := range test.set {
			os.Setenv(k, v)
		}

		res, err := configApp()

		if !assert.Equal(t, test.err, err, "%v: Expect: %v, Got: %v", test.name, test.err, err) ||
			!assert.Equal(t, test.res, res, "%v: Expect: %v, Got: %v", test.name, test.res, res) {
			t.Fatal()
		}
	}
}
