package fstopimp

import (
	"fmt"
	"github.com/RangelReale/filesharetop/lib"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"strconv"
	"strings"
	"time"
)

type Importer struct {
	logger          *log.Logger
	Session         *mgo.Session
	Database        string
	ScoreCalculator ScoreCalculator
}

func BuildCurrentCollectionName(id string) string {
	cconsname := "current"
	if id != "" {
		cconsname = fmt.Sprintf("%s_%s", cconsname, id)
	}
	return cconsname
}

func NewImporter(logger *log.Logger, session *mgo.Session) *Importer {
	return &Importer{
		logger:          logger,
		Session:         session,
		Database:        "filesharetop",
		ScoreCalculator: &DefaultScoreCalculator{},
	}
}

func (i *Importer) Import(fetcher fstoplib.Fetcher) error {
	fetcher.SetLogger(i.logger)

	list, err := fetcher.Fetch()
	if err != nil {
		return err
	}

	dt := time.Now().UTC()

	rec := &FSTopRecord{
		Date:       dt.Format("2006-01-02"),
		Hour:       dt.Hour(),
		ImportTime: dt,
		List:       list,
	}

	c := i.Session.DB(i.Database).C("history")
	ccat := i.Session.DB(i.Database).C("category")

	_, err = c.Upsert(bson.M{
		"date": rec.Date,
		"hour": rec.Hour,
	}, rec)
	if err != nil {
		return err
	}

	// import categories
	_, err = ccat.RemoveAll(nil)
	if err != nil {
		return err
	}

	fcat, err := fetcher.CategoryMap()
	if err != nil {
		return err
	}

	for ch, ci := range *fcat {
		err = ccat.Insert(&FSSTopCategory{
			ID:   ch,
			List: ci,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Importer) Consolidate(id string, hours int) error {
	c := i.Session.DB(i.Database).C("history")
	c.EnsureIndexKey("date", "hour")

	ccons := i.Session.DB(i.Database).C(BuildCurrentCollectionName(id))

	dtlim := time.Now().UTC().Add(-1 * (time.Hour * time.Duration(hours)))

	//i.logger.Printf("Start date: %s\n", dtlim.Format("2006-01-02"))

	iter := c.Find(bson.M{
		"$or": []bson.M{
			{
				"date": bson.M{"$gte": dtlim.Format("2006-01-02")},
				"hour": bson.M{"$gte": strconv.Itoa(dtlim.Hour())},
			},
			{
				"date": bson.M{"$gt": dtlim.Format("2006-01-02")},
			},
		}}).Sort("date", "hour").Iter()

	items := make(map[string]*FSTopStats)
	var rec FSTopRecord

	cttotal := int32(0)
	for iter.Next(&rec) {
		//i.logger.Printf("%s - %s\n", rec.Date, rec.Hour)
		cttotal++

		for _, pi := range rec.List {
			var item *FSTopStats
			var ok bool

			if item, ok = items[pi.Id]; !ok {
				item = &FSTopStats{
					Id:       pi.Id,
					Title:    strings.TrimSpace(pi.Title),
					Link:     pi.Link,
					Category: pi.Category,
				}
				items[pi.Id] = item
			}

			if item.Last != nil {
				//i.logger.Printf("%s - [%d] [%d] [%d]\n", item.Title, pi.Seeders-item.Last.Seeders,
				//pi.Leechers-item.Last.Leechers, pi.Complete-item.Last.Complete)

				item.Score += i.ScoreCalculator.CalcScore(rec.Date, rec.Hour, pi, item.Last)
			}
			item.Count++
			item.Last = pi
		}
	}

	var err error
	if err = iter.Close(); err != nil {
		return err
	}

	// clear "current" collection
	_, err = ccons.RemoveAll(nil)
	if err != nil {
		return err
	}

	// insert items in current collection
	for _, ii := range items {
		err = ccons.Insert(ii)
		if err != nil {
			return err
		}
	}

	return nil
}
