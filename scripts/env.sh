#!/bin/bash

export PGPASSWORD='pass123'
export DB_USER_ROOT='postgres'
export DB_HOST=$(hostname -I | awk '{print $1}')
export DB_TYPE='postgres'
export DB_SSLMODE=disable
export DB_PORT=5432
export APP_PORT=':8080'

export DB_USER='order_service'
export DB_NAME='order_service'
export DB_PASSWORD='mik9iaLeexaer9vi4eew9E'

export KAFKA_HOST=$(hostname -I | awk '{print $1}')
export KAFKA_PORT_1=9091
export KAFKA_PORT_2=9092
export KAFKA_PORT_3=9093
export KAFKA_TOPIC='order'
export KAFKA_GROUP_NAME='order-group'

export FRONT_HOST=localhost
export FRONT_PORT=8081

