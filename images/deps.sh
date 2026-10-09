
set -xeuo pipefail

# CockroachDB

wget -qO- https://binaries.cockroachdb.com/cockroach-v23.2.3.linux-amd64.tgz | tar xvz
mv cockroach-v23.2.3.linux-amd64/cockroach /usr/local/bin/
mv cockroach-v23.2.3.linux-amd64/lib/* /usr/lib/

# Postgres & Redis

apt install curl ca-certificates
install -d /usr/share/postgresql-common/pgdg
curl -o /usr/share/postgresql-common/pgdg/apt.postgresql.org.asc --fail https://www.postgresql.org/media/keys/ACCC4CF8.asc

# Create the repository configuration file:
. /etc/os-release
sh -c "echo 'deb [signed-by=/usr/share/postgresql-common/pgdg/apt.postgresql.org.asc] https://apt.postgresql.org/pub/repos/apt $VERSION_CODENAME-pgdg main' > /etc/apt/sources.list.d/pgdg.list"

# Update the package lists:
apt update && apt install -y -qq postgresql-17 redis-server

rm /etc/postgresql/17/main/pg_hba.conf
echo 'local   all             all                                     trust' >> /etc/postgresql/17/main/pg_hba.conf
echo 'host    all             all             127.0.0.1/8             trust' >> /etc/postgresql/17/main/pg_hba.conf
echo 'host    all             all             ::1/128                 trust' >> /etc/postgresql/17/main/pg_hba.conf
echo 'host    all             all             ::0/0                   trust' >> /etc/postgresql/17/main/pg_hba.conf

echo 'max_connections = 1000' >> /etc/postgresql/17/main/conf.d/connectionlimits.conf
echo 'fsync = off' >> /etc/postgresql/17/main/conf.d/nosync.conf

# TiDB local test cluster helper
TIDB_VERSION=v8.5.6
curl --proto '=https' --tlsv1.2 -sSf https://tiup-mirrors.pingcap.com/install.sh | sh
ln -s ~/.tiup/bin/tiup /usr/local/bin/tiup
PLAYGROUND_VERSION=v1.16.5
tiup install playground:${PLAYGROUND_VERSION} \
    tidb:${TIDB_VERSION} \
    tikv:${TIDB_VERSION} \
    pd:${TIDB_VERSION} \
    tiflash:${TIDB_VERSION}
