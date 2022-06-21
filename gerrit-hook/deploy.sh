#!/usr/bin/env bash
set -euxo pipefail
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
gcloud compute scp ./gerrit-hook  --zone "us-central1-a" gerrit@gerrit:  --tunnel-through-iap --project "storj-developer-team"
