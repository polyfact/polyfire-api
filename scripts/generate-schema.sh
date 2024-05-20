#!/bin/bash

TMP_DB_PORT=5433

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

DATA_PG_DIR=$(mktemp)
rm $DATA_PG_DIR && mkdir $DATA_PG_DIR
initdb $DATA_PG_DIR -U postgres 1>&2 || exit 1

SOCKET_DIR=$(mktemp)
rm $SOCKET_DIR && mkdir -p $SOCKET_DIR/postgresql
echo "unix_socket_directories = '$SOCKET_DIR/postgresql'" >> $DATA_PG_DIR/postgresql.conf

postgres -D $DATA_PG_DIR -p $TMP_DB_PORT 1>&2 &

POSTGRES_PID=$!

cleanup() {
	kill $POSTGRES_PID
	wait $POSTGRES_PID

	rm -rf $DATA_PG_DIR $SOCKET_DIR
}

error() {
	cleanup && exit 1
}

POSTGRES_URI="postgres://postgres:postgres@localhost:$TMP_DB_PORT/postgres"

POSTGRES_URI=$POSTGRES_URI python $SCRIPT_DIR/migrate.py init 1>&2 || error && POSTGRES_URI=$POSTGRES_URI python $SCRIPT_DIR/migrate.py migrate --ignore-rls 1>&2 || error

pg_dump -s "postgres://postgres:postgres@localhost:$TMP_DB_PORT/postgres" || error

cleanup

