#!/bin/sh

go-bindata -pkg fstopsite -o res.go res/...
