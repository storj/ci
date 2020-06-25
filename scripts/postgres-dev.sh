#!/usr/bin/env bash

LOG_FILE=${STORJ_SIM_POSTGRES_LOG:-"storj-sim-postgres.log"}
CONTAINER_NAME=storj_sim_postgres

cleanup(){
  docker rm -f $CONTAINER_NAME
}
trap cleanup EXIT

docker run --rm -d -p 5433:5432 --env POSTGRES_HOST_AUTH_METHOD=trust --name $CONTAINER_NAME postgres:12.3 -c log_min_duration_statement=0
docker logs -f $CONTAINER_NAME > $LOG_FILE 2>&1 &

STORJ_SIM_DATABASE=${STORJ_SIM_DATABASE:-"teststorj"}

RETRIES=10

until docker exec $CONTAINER_NAME psql -h localhost -U postgres -d postgres -c "select 1" > /dev/null 2>&1 || [ $RETRIES -eq 0 ]; do
  echo "Waiting for postgres server, $((RETRIES--)) remaining attempts..."
  sleep 1
done

docker exec $CONTAINER_NAME psql -h localhost -U postgres -c "create database $STORJ_SIM_DATABASE;"

export STORJ_SIM_POSTGRES="postgres://postgres@localhost:5433/$STORJ_SIM_DATABASE?sslmode=disable"