#!/bin/bash
SVC="http://localhost:8090"
if [ -z "$1" ]; then
  LONG="https://surrealdb.com"
else
  LONG="$1"
fi
curl -X POST --data "url=$LONG" $SVC/shorten
