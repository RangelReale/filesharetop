package fstoplib

import (
	"log"
)

type Fetcher interface {
	ID() string
	SetLogger(l *log.Logger)
	Fetch() (map[string]*Item, error)
	CategoryMap() (*CategoryMap, error)
}
