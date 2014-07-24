#!/bin/bash

cd server ; go build -o ../server-app ; cd ..
cd first ; go build -o ../first-app ; cd ..
cd second ; go build -o ../second-app ; cd ..
cd third ; go build -o ../third-app ; cd ..

./first-app &
./second-app &
./third-app &
./server-app

killall server-app
killall first-app
killall second-app
killall third-app

rm server-app
rm first-app
rm second-app
rm third-app
