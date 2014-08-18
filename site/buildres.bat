@echo off

rem go-bindata -debug -pkg fstopsite -o res.go res/...
go-bindata -pkg fstopsite -o res.go res/...

echo Finished

rem pause
