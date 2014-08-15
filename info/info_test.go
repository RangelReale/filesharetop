package fstopinfo

import (
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
		log.Printf("Imported %d\n", len(d))

		for _, ii := range d {
			log.Printf("T: %s - Score %d\n", ii.Title, ii.Score)
		}
	*/
}
