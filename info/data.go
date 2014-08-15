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
