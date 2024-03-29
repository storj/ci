set -xeuo pipefail

# CockroachDB

wget -qO- https://binaries.cockroachdb.com/cockroach-v23.2.2.linux-amd64.tgz | tar xvz
mv cockroach-v23.2.2.linux-amd64/cockroach /usr/local/bin/
mv cockroach-v23.2.2.linux-amd64/lib/* /usr/lib/

# Postgres

curl -sf https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
echo "deb http://apt.postgresql.org/pub/repos/apt/ bookworm-pgdg main" | tee /etc/apt/sources.list.d/pgdg.list

apt-get update && apt-get install -y -qq postgresql-13 redis-server

rm /etc/postgresql/13/main/pg_hba.conf
echo 'local   all             all                                     trust' >> /etc/postgresql/13/main/pg_hba.conf
echo 'host    all             all             127.0.0.1/8             trust' >> /etc/postgresql/13/main/pg_hba.conf
echo 'host    all             all             ::1/128                 trust' >> /etc/postgresql/13/main/pg_hba.conf
echo 'host    all             all             ::0/0                   trust' >> /etc/postgresql/13/main/pg_hba.conf

echo 'max_connections = 1000' >> /etc/postgresql/13/main/conf.d/connectionlimits.conf
echo 'fsync = off' >> /etc/postgresql/13/main/conf.d/nosync.conf
