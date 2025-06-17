#!/usr/bin/env bash
BIN=shard-server
GOOS=linux GOARCH=amd64 go build -o $BIN shard-server.go

for HOST in 13.56.233.224 54.193.44.210 50.18.38.42; do
  scp -i rith-key.pem $BIN ubuntu@$HOST:~/
  ssh -i rith-key.pem ubuntu@$HOST "nohup ~/shard-server &> shard-server.log &"
done