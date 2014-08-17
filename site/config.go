package fstopsite

import (
	"gopkg.in/mgo.v2"
	"log"
)

type Config struct {
	Title       string
	Port        int
	Session     *mgo.Session
	Logger      *log.Logger
	Database    string
	TopId       string
	HistoryDays int
}

func NewConfig(port int) *Config {
	return &Config{
		Port:        port,
		Database:    "filesharetop",
		HistoryDays: 168,
	}
}
