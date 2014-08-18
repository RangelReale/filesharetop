package fstopinfo

import (
	"github.com/RangelReale/filesharetop/lib"
	"time"
)

type FSInfoHistory struct {
	Date       string
	Hour       int
	ImportTime time.Time
	Item       *fstoplib.Item
}

type FSCategoryList []string

func (c FSCategoryList) Exists(category string) bool {
	for _, cc := range c {
		if cc == category {
			return true
		}
	}
	return false
}
