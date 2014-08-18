package fstopinfo

import (
	"errors"
	"fmt"
	"github.com/RangelReale/filesharetop/importer"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"strconv"
	"time"
)

type Info struct {
	logger   *log.Logger
	Session  *mgo.Session
	Database string
}

func NewInfo(logger *log.Logger, session *mgo.Session) *Info {
	return &Info{
		logger:   logger,
		Session:  session,
		Database: "filesharetop",
	}
}

func (i *Info) Top(id string) ([]*fstopimp.FSTopStats, error) {
	ccons := i.Session.DB(i.Database).C(fstopimp.BuildCurrentCollectionName(id))
	ccons.EnsureIndexKey("-score")

	items := make([]*fstopimp.FSTopStats, 0)

	iter := ccons.Find(nil).Sort("-score").Iter()
	var stats fstopimp.FSTopStats

	for iter.Next(&stats) {
		s := stats
		items = append(items, &s)
	}
	var err error
	if err = iter.Close(); err != nil {
		return nil, err
	}
	return items, nil
}

func (i *Info) TopCategory(id string, category string) ([]*fstopimp.FSTopStats, error) {
	ccat := i.Session.DB(i.Database).C("category")
	ccat.EnsureIndexKey("id")

	var catinfo fstopimp.FSSTopCategory
	err := ccat.Find(bson.M{"id": category}).One(&catinfo)
	if err == mgo.ErrNotFound {
		return nil, errors.New("Category not found")
	} else if err != nil {
		return nil, err
	}

	ccons := i.Session.DB(i.Database).C(fstopimp.BuildCurrentCollectionName(id))
	ccons.EnsureIndexKey("-score")

	items := make([]*fstopimp.FSTopStats, 0)

	iter := ccons.Find(nil).Sort("-score").Iter()
	var stats fstopimp.FSTopStats

	for iter.Next(&stats) {
		if catinfo.IsContained(stats.Category) {
			s := stats
			items = append(items, &s)
		}
	}
	if err = iter.Close(); err != nil {
		return nil, err
	}
	return items, nil
}

func (i *Info) History(id string, hours int) ([]*FSInfoHistory, error) {
	c := i.Session.DB(i.Database).C("history")
	c.EnsureIndexKey("date", "hour")

	dtlim := time.Now().UTC().Add(-1 * (time.Hour * time.Duration(hours)))

	iter := c.Find(bson.M{
		"$or": []bson.M{
			{
				"date": bson.M{"$gte": dtlim.Format("2006-01-02")},
				"hour": bson.M{"$gte": strconv.Itoa(dtlim.Hour())},
			},
			{
				"date": bson.M{"$gt": dtlim.Format("2006-01-02")},
			},
		}}).Select(bson.M{"date": 1, "hour": 1, "import_time": 1, fmt.Sprintf("list.%s", id): 1}).Sort("date", "hour").Iter()

	items := make([]*FSInfoHistory, 0)
	found := false
	var rec fstopimp.FSTopRecord
	for iter.Next(&rec) {
		ni := &FSInfoHistory{
			Date:       rec.Date,
			Hour:       rec.Hour,
			ImportTime: rec.ImportTime,
		}
		if ri, ok := rec.List[id]; ok {
			ni.Item = ri
			found = true
		}
		items = append(items, ni)
	}
	var err error
	if err = iter.Close(); err != nil {
		return nil, err
	}

	if !found {
		return nil, nil
	}

	return items, nil
}

func (i *Info) Categories() (FSCategoryList, error) {
	c := i.Session.DB(i.Database).C("category")
	c.EnsureIndexKey("id")

	iter := c.Find(nil).Select(bson.M{"id": 1}).Sort("id").Iter()

	items := make(FSCategoryList, 0)
	var rec fstopimp.FSSTopCategory
	for iter.Next(&rec) {
		items = append(items, rec.ID)
	}
	var err error
	if err = iter.Close(); err != nil {
		return nil, err
	}
	return items, nil
}
