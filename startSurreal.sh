#!/bin/bash
OVERRIDE_ADDRESS="-b 0.0.0.0:8000" # Optional, to change the bind port from 8000
FORCE_SCHEMA="--strict" # Optional, to prevent new fields being created in each record
RUN_DB_WITH_ADMIN="--user root --pass root" # Optional, to configure database before exposing

# Modes of operation; Pick one
IN_MEMORY="memory" # Default mode of operation
ROCKSDB="file://$HOME/urlshortener-demo" # If you want to persist data
TIKV="tikv://address-to-tikv-cluster:2379" # For when you want to scale your storage layer

surreal start $OVERRIDE_ADDRESS $FORCE_SCHEMA $IN_MEMORY

