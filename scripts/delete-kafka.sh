#!/bin/bash

docker compose -f $(pwd)/docker-compose-kafka.yaml down
docker volume rm $(docker volume ls -q)
