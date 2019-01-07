package app

import (
	"net/url"
	"strconv"
	"strings"
)

type Config struct {
	secretKey      string
	dataSourceName string
}

func NewConfig(secretKey string, dataSourceName string) *Config {
	return &Config{secretKey: secretKey, dataSourceName: dataSourceName}
}

func (c *Config) String() string {
	var safeDSN string
	u, err := url.Parse(c.dataSourceName)
	if err != nil {
		safeDSN = "InvalidDSN"
	}
	if pass, passSet := u.User.Password(); passSet {
		safeDSN = strings.Replace(u.String(), pass+"@", "*****@", 1)
	}

	var s strings.Builder
	s.WriteString("Config{")
	s.WriteString("len(secretKey)=" + strconv.Itoa(len(c.secretKey)))
	s.WriteString(" dataSourceName=" + safeDSN)
	s.WriteString("}")
	return s.String()
}
