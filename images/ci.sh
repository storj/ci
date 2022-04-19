set -euo pipefail

bash ./scripts/install-awscli.sh
export PATH="$PATH:/root/bin"

# Android/Java binding tests
apt-get install -y default-jre

## Tools for gateway testing
apt-get install -y s3fs
# Duplicity backup tool for S3 gateway test scenarios
apt-get install -y duplicity python3-pip && pip install boto3

# rclone and test tool for S3 gateway test scenarios
go install github.com/rclone/rclone@v1.58.0
go install github.com/rclone/rclone/fstest/test_all@v1.58.0

# Duplicati backup tool for S3 gateway test scenarios
apt-get -y install mono-devel
curl -sfL https://github.com/duplicati/duplicati/releases/download/v2.0.5.114-2.0.5.114_canary_2021-03-10/duplicati_2.0.5.114-1_all.deb -o /tmp/duplicati.deb
apt -y install /tmp/duplicati.deb

# Requirements for UI tests
DEBIAN_FRONTEND="noninteractive" apt install -y chromium xorg xvfb gtk2-engines-pixbuf dbus-x11 xfonts-base xfonts-100dpi xfonts-75dpi xfonts-cyrillic xfonts-scalable imagemagick x11-apps
