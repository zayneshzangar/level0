#!/bin/bash
source $(pwd)/env.sh

docker compose -f $(pwd)/docker-compose-psql.yaml up -d
