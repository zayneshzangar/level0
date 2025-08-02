#!/bin/bash
source $(pwd)/env.sh

docker compose -f $(pwd)/docker-compose-kafka.yaml up -d
