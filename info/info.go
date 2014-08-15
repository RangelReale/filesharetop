package fstopinfo

import (
	//"fmt"
	"github.com/RangelReale/filesharetop/importer"
	"gopkg.in/mgo.v2"
	//"gopkg.in/mgo.v2/bson"
	"log"
	//"sort"
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
