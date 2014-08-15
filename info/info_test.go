package fstopinfo

import (
	//"fmt"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"log"
	"testing"
)

func TestInfo(t *testing.T) {
	// connect to mongodb
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Panic(err)
	}
	defer session.Close()

	i := NewInfo(log.New(ioutil.Discard, "", 0), session)

	_, err = i.Top()
	if err != nil {
		t.Error(err.Error())
		return
	}

	/*
		log.Printf("Found %d\n", len(d))

		for _, ii := range d {
			log.Printf("T: %s - Score %d\n", ii.Title, ii.Score)
		}
	*/
}

func TestInfoCategory(t *testing.T) {
	// connect to mongodb
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Panic(err)
	}
	defer session.Close()

	i := NewInfo(log.New(ioutil.Discard, "", 0), session)

	_, err = i.TopCategory("MOVIE")
	if err != nil {
		t.Error(err.Error())
		return
	}

	/*
		log.Printf("Found %d\n", len(d))

		for _, ii := range d {
			log.Printf("T: %s - Score %d\n", ii.Title, ii.Score)
		}
	*/
}

func TestInfoHistory(t *testing.T) {
	// connect to mongodb
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Panic(err)
	}
	defer session.Close()

	i := NewInfo(log.New(ioutil.Discard, "", 0), session)

	d, err := i.History("e45bfaeddc4a2c0475e4479b67a9c2e6e29863fd", 48)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if d == nil {
		t.Error("Not found")
	}

	/*
		for _, ii := range d {
			ls := fmt.Sprintf("%s - %d", ii.Date, ii.Hour)
			if ii.Item != nil {
				ls = fmt.Sprintf("%s: %s [%d]", ls, ii.Item.Title, ii.Item.Seeders)
			}
			log.Println(ls)
		}
	*/
}
