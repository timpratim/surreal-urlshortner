#!/bin/bash
SVC="http://localhost:8090"
if [ -z "$1" ]; then
  LONG="https://www.youtube.com/watch?v=C7WFwgDRStM&t=97s"
else
  LONG="$1"
fi
curl -X POST --data "url=$LONG" $SVC/shorten
