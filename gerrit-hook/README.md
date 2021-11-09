# General gerrit-hook binary

Gerrit-hook can handle any type of gerrit hooks.

To install create a symlink from the binary to the `$GERRIT_SITE/hooks`:

```bash
gcloud beta compute ssh --zone "us-central1-a" "gerrit"  --tunnel-through-iap --project "storj-developer-team"
sudo su - gerrit
cd /home/gerrit/site/hooks
ln -s /home/gerrit/gerrit-hook patchset-created
```

The hook will be executed for all `patchset-created` event for `ALL` repository.

Example of parameters used by gerrit:

```
--change storj%2Fup~main~I684af6cd8a0c49baa7b55c2298fbd1974f5c56fe --kind REWORK --change-url https://review.dev.storj.io/c/storj/up/+/6241 --change-owner "Elek, Márton <marton@storj.io>" --change-owner-username elek --project storj/up --branch main --topic  --uploader "Elek, Márton <marton@storj.io>" --uploader-username elek --commit c72759e5db315f1cdbe9c8529cecc003f2c38f3e --patchset 3gerrit@gerrit
```

The current implementation updates the linked Github issues with the Github token stored
in `/home/gerrit/.config/gerrit-hook/github-token`