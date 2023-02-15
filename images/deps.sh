set -euo pipefail

# CockroachDB

wget -qO- https://binaries.cockroachdb.com/cockroach-v22.2.5.linux-amd64.tgz | tar xvz
cp -i cockroach-v22.2.5.linux-amd64/cockroach /usr/local/bin/

# Postgres

curl -sf https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
echo "deb http://apt.postgresql.org/pub/repos/apt/ bullseye-pgdg main" | tee /etc/apt/sources.list.d/pgdg.list

apt-get update && apt-get install -y -qq postgresql-13 redis-server

rm /etc/postgresql/13/main/pg_hba.conf
echo 'local   all             all                                     trust' >> /etc/postgresql/13/main/pg_hba.conf
echo 'host    all             all             127.0.0.1/8             trust' >> /etc/postgresql/13/main/pg_hba.conf
echo 'host    all             all             ::1/128                 trust' >> /etc/postgresql/13/main/pg_hba.conf
echo 'host    all             all            ::0/0                   trust' >> /etc/postgresql/13/main/pg_hba.conf

echo 'max_connections = 1000' >> /etc/postgresql/13/main/conf.d/connectionlimits.conf
echo 'fsync = off' >> /etc/postgresql/13/main/conf.d/nosync.conf
