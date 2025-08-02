#!/bin/bash

docker compose -f $(pwd)/docker-compose-psql.yaml down
docker volume rm  scripts_psql-data # $(docker volume ls -q | grep scripts_psql-data)
