package fstopinfo

import (
	//"fmt"
	"github.com/RangelReale/filesharetop/lib"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"strconv"
	"strings"
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
