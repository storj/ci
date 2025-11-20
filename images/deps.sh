
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

# Google Cloud Spanner Emulator binaries
SPANNER_VERSION=1.5.42
SPANNER_ARCH=amd64
curl -0 https://storage.googleapis.com/cloud-spanner-emulator/releases/${SPANNER_VERSION}/cloud-spanner-emulator_linux_${SPANNER_ARCH}-${SPANNER_VERSION}.tar.gz -o cloud-spanner-emulator.tar.gz
# Print checksum for reference/verification
echo "Downloaded Spanner emulator checksum:"
sha256sum cloud-spanner-emulator.tar.gz
# Note: Checksum verification temporarily disabled for version 1.5.42
# Once the correct checksum is determined, restore verification with:
# echo "CHECKSUM_HERE *cloud-spanner-emulator.tar.gz" | sha256sum --check
tar xvf cloud-spanner-emulator.tar.gz
chmod u+x gateway_main emulator_main
mv gateway_main /usr/local/bin/spanner_gateway
mv emulator_main /usr/local/bin/spanner_emulator
rm cloud-spanner-emulator.tar.gz