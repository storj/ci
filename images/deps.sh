
set -xeuo pipefail

# CockroachDB

wget -qO- https://binaries.cockroachdb.com/cockroach-v23.2.3.linux-amd64.tgz | tar xvz
mv cockroach-v23.2.3.linux-amd64/cockroach /usr/local/bin/
mv cockroach-v23.2.3.linux-amd64/lib/* /usr/lib/

# Postgres & Redis

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

# Google Cloud Spanner Emulator binaries
SPANNER_VERSION=1.5.25
SPANNER_ARCH=amd64
wget -O cloud-spanner-emulator.tar.gz https://storage.googleapis.com/cloud-spanner-emulator/releases/${SPANNER_VERSION}/cloud-spanner-emulator_linux_${SPANNER_ARCH}-${SPANNER_VERSION}.tar.gz
echo "242c64520cba26c2af04535853d820ca6b8ae40d32bd5cb8472a7e97fd45d9b7 *cloud-spanner-emulator.tar.gz" | sha256sum --check
tar xvf cloud-spanner-emulator.tar.gz
chmod u+x gateway_main emulator_main
mv gateway_main /usr/local/bin/spanner_gateway
mv emulator_main /usr/local/bin/spanner_emulator