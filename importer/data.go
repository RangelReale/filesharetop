package fstopimp

import (
	"github.com/RangelReale/filesharetop/lib"
	"time"
)

type FSTopRecord struct {
	Date       string                    `bson:"date"`
	Hour       int                       `bson:"hour"`
	ImportTime time.Time                 `bson:"import_time"`
	List       map[string]*fstoplib.Item `bson:"list"`
}

type FSSTopCategory struct {
	ID   string   `bson:"id"`
	List []string `bson:"list"`
}

func (c *FSSTopCategory) IsContained(category string) bool {
	for _, v := range c.List {
		if v == category {
			return true
		}
	}
	return false
}

type FSTopStats struct {
	Id       string `bson:"id"`
	Title    string `bson:"title"`
	Link     string `bson:"link"`
	Category string `bson:"category"`
	Count    int32  `bson:"count"`
	Score    int32  `bson:"score"`

	Last *fstoplib.Item `bson:"last"`
}

type FSTopStatsSorted []*FSTopStats

func (d FSTopStatsSorted) Len() int {
	return len(d)
}

func (d FSTopStatsSorted) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// Inverted to sort in descent order
func (d FSTopStatsSorted) Less(i, j int) bool {
	return d[i].Score > d[j].Score
}
