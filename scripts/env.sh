#!/bin/bash

export PGPASSWORD='pass123'
export DB_USER_ROOT='postgres'
export DB_HOST=$(hostname -I | awk '{print $1}')
export DB_TYPE='postgres'
export DB_SSLMODE=disable
export DB_PORT=5432
export APP_PORT=8080

export DB_USER='order_service'
export DB_NAME='order_service'
export DB_PASSWORD='mik9iaLeexaer9vi4eew9E'
