#!/bin/bash
SVC="http://localhost:8090"
if [ -z "$1" ]; then
  echo "Must provide a url"
  exit 1
else
  LONG="$1"
fi
curl -X GET $LONG
