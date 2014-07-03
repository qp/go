#!/bin/bash

go build server.go
go build first.go
go build second.go
go build third.go

./first &
./second &
./third &
./server

rm server
rm first
rm second
rm third
