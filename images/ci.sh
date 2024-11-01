set -xeuo pipefail

bash ./scripts/install-awscli.sh

## Tools for gateway testing
apt-get install -y s3fs
# Duplicity backup tool for S3 gateway test scenarios
apt-get install -y duplicity python3-pip python3-boto3
# Tool for running github.com/ceph/s3-tests for gateway
apt-get install -y tox

# Duplicati backup tool for S3 gateway test scenarios
apt-get -y install mono-devel
curl -sfL https://updates.duplicati.com/beta/duplicati_2.0.7.1-1_all.deb -o /tmp/duplicati.deb
apt -y install /tmp/duplicati.deb

# Requirements for UI tests
npm install @playwright/test@v1.48.2
npx playwright install --with-deps