source $(pwd)/env.sh

psql -U $DB_USER_ROOT -p $DB_PORT -h $DB_HOST \
    -c "CREATE ROLE $DB_USER WITH PASSWORD '$DB_PASSWORD';" || true
psql -U $DB_USER_ROOT -p $DB_PORT -h $DB_HOST \
    -c "CREATE DATABASE $DB_NAME;" || true
psql -U $DB_USER_ROOT -p $DB_PORT -h $DB_HOST \
    -c "ALTER ROLE $DB_USER WITH LOGIN;" || true
psql -U $DB_USER_ROOT -p $DB_PORT -h $DB_HOST -d $DB_NAME \
    -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;" || true
psql -U $DB_USER_ROOT -p $DB_PORT -h $DB_HOST -d $DB_NAME \
    -c "GRANT CREATE ON SCHEMA public TO $DB_USER;" || true
# Предоставление прав на все таблицы в схеме public (для будущих таблиц)
psql -U $DB_USER_ROOT -p $DB_PORT -h $DB_HOST -d $DB_NAME \
    -c "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;" || true
