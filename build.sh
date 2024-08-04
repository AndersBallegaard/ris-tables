#!/bin/bash

for arch in linux-amd64 windows-amd64 darwin-arm64
do
GOOS=$(echo $arch | cut -d "-" -f1)
GOARCH=$(echo $arch | cut -d "-" -f2)
EXTENTION=""
if [ "$GOOS" == "windows" ]
then 
EXTENTION=".exe"
fi
go build -o bin/ris_tables-$GOOS-$GOARCH$EXTENTION
done