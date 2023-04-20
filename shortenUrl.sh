#!/bin/bash
SVC="http://localhost:8090"
LONG="https://surrealdb.com"
curl -X POST --data "{\"url\": \"$LONG\"}" $SVC/shorten
