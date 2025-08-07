#!/bin/bash


if [[ "$1" == "--delete" ]]; then
    echo "Удаление базы данных и брокеров сообщении..."
    source $(pwd)/delete-psql.sh
    sleep 5
    source $(pwd)/delete-kafka.sh
else
    echo "Создание и запуск базы данных и брокеров сообщении..."
    $(pwd)/run-psql.sh
    sleep 5
    $(pwd)/create-db.sh
    $(pwd)/run-kafka.sh
fi
