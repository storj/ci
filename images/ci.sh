set -xeuo pipefail

## Tools for gateway testing
apt-get install -y s3fs awscli
# Duplicity backup tool for S3 gateway test scenarios
apt-get install -y duplicity python3-pip python3-boto3
# Tool for running github.com/ceph/s3-tests for gateway
apt-get install -y tox

# Duplicati backup tool for S3 gateway test scenarios
apt-get -y install mono-devel libicu76
curl -sfL https://updates.duplicati.com/canary/duplicati-2.1.1.100_canary_2025-08-08-linux-x64-cli.deb -o /tmp/duplicati.deb
apt -y install /tmp/duplicati.deb

# Requirements for UI tests
npm install @playwright/test@next # TODO use v1.55.0 once it's released
npx playwright install --with-deps

# Install Zig for cross-compiling
apt-get install -y xz-utils
VERSION=0.14.0
wget https://ziglang.org/download/$VERSION/zig-linux-x86_64-$VERSION.tar.xz
tar -xJf zig-linux-x86_64-$VERSION.tar.xz
rm zig-linux-x86_64-$VERSION.tar.xz
mv zig-linux-x86_64-$VERSION /usr/local/zig
