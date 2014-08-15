package fstopinfo

import (
	//"fmt"
	"errors"
	"github.com/RangelReale/filesharetop/importer"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
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

func (i *Info) Top() ([]*fstopimp.FSTopStats, error) {
	ccons := i.Session.DB(i.Database).C("current")
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

func (i *Info) TopCategory(category string) ([]*fstopimp.FSTopStats, error) {
	ccat := i.Session.DB(i.Database).C("category")
	ccat.EnsureIndexKey("id")

	var catinfo fstopimp.FSSTopCategory
	err := ccat.Find(bson.M{"id": category}).One(&catinfo)
	if err == mgo.ErrNotFound {
		return nil, errors.New("Category not found")
	} else if err != nil {
		return nil, err
	}

	ccons := i.Session.DB(i.Database).C("current")
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
