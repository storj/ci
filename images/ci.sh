set -xeuo pipefail

# AWS CLI v2.22.35 is the latest version we support currently.
# This should be unpegged once https://github.com/storj/gateway-st/issues/89 is solved.
curl https://awscli.amazonaws.com/awscli-exe-linux-x86_64-2.22.35.zip -o awscliv2.zip
unzip -q awscliv2.zip
./aws/install --bin-dir /usr/local/bin --install-dir /usr/local/aws-cli
rm awscliv2.zip
rm -r ./aws

## Tools for gateway testing
apt-get install -y s3fs
# Duplicity backup tool for S3 gateway test scenarios
apt-get install -y duplicity python3-pip python3-boto3
# Tool for running github.com/ceph/s3-tests for gateway
apt-get install -y tox

# Duplicati backup tool for S3 gateway test scenarios
DUPLICATI_NAME=duplicati-2.1.0.5_stable_2025-03-04-linux-x64-cli
curl https://updates.duplicati.com/stable/${DUPLICATI_NAME}.zip -o duplicati.zip
unzip -q duplicati.zip
rm duplicati.zip
mv "${DUPLICATI_NAME}" /usr/local/duplicati
ln -s /usr/local/duplicati/duplicati-cli /usr/local/bin/duplicati-cli

# Requirements for UI tests
npm install @playwright/test@v1.56.0
npx playwright install --with-deps

# Install Zig for cross-compiling
apt-get install -y xz-utils
VERSION=0.14.0
wget https://ziglang.org/download/$VERSION/zig-linux-x86_64-$VERSION.tar.xz
tar -xJf zig-linux-x86_64-$VERSION.tar.xz
rm zig-linux-x86_64-$VERSION.tar.xz
mv zig-linux-x86_64-$VERSION /usr/local/zig
