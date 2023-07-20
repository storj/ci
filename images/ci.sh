set -xeuo pipefail

bash ./scripts/install-awscli.sh

# Android/Java binding tests
apt-get install -y default-jre

## Tools for gateway testing
apt-get install -y s3fs
# Duplicity backup tool for S3 gateway test scenarios
apt-get install -y duplicity python3-pip python3-boto3

# rclone and test tool for S3 gateway test scenarios
go install github.com/rclone/rclone@v1.58.0
go install github.com/rclone/rclone/fstest/test_all@v1.58.0

# Duplicati backup tool for S3 gateway test scenarios
apt-get -y install mono-devel
curl -sfL https://updates.duplicati.com/beta/duplicati_2.0.7.1-1_all.deb -o /tmp/duplicati.deb
apt -y install /tmp/duplicati.deb

# Requirements for UI tests
npx playwright install-deps