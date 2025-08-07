#!/bin/bash

docker exec -it kafka-1 bash -c \
    "kafka-topics --create --bootstrap-server kafka-1:29091 --replication-factor 1 --partitions 1 --topic order"
