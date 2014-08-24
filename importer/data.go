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
	Id               string `bson:"id"`
	Title            string `bson:"title"`
	Link             string `bson:"link"`
	Category         string `bson:"category"`
	Count            int32  `bson:"count"`
	Score            int32  `bson:"score"`
	FirstAppearCount int32  `bson:"-"`

	Last *fstoplib.Item `bson:"last"`
}

type FSTopStatsList []*FSTopStats

func (l FSTopStatsList) Paged(page int, pagesize int) FSTopStatsList {
	ret := make(FSTopStatsList, 0, pagesize)
	if page < 1 {
		page = 1
	}
	start := (page - 1) * pagesize
	end := page * pagesize

	for i := start; i < end; i++ {
		if i > len(l)-1 {
			break
		}
		ret = append(ret, l[i])
	}

	return ret
}

func (l FSTopStatsList) PageCount(pagesize int) int {
	ret := len(l) / pagesize
	if len(l)%pagesize != 0 {
		ret++
	}
	return ret
}
