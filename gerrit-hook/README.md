# General gerrit-hook binary

Gerrit-hook can handle any type of gerrit hooks.

To install create a symlink from the binary to the `$GERRIT_SITE/hooks` to the compiled binary.

In Storj labs environment it can be installed by `./deploy.sh`.

Example of parameters used by gerrit:

```
--change storj%2Fup~main~I684af6cd8a0c49baa7b55c2298fbd1974f5c56fe --kind REWORK --change-url https://review.dev.storj.io/c/storj/up/+/6241 --change-owner "Elek, Márton <marton@storj.io>" --change-owner-username elek --project storj/up --branch main --topic  --uploader "Elek, Márton <marton@storj.io>" --uploader-username elek --commit c72759e5db315f1cdbe9c8529cecc003f2c38f3e --patchset 3gerrit@gerrit
```

The current implementation 
 * updates the linked Github issues with the Github token stored
 * Trigger jenkins build for some comments (-verify / premerge)

in `/home/gerrit/.config/gerrit-hook/config.yaml`

## Development

In the final deployment the action is selected by the binary name (as the binary is symlinked to the hooks directory)

For local development:

 1. Create `/tmp/gerrit-hook-debug/` directory on the gerrit server.
 2. Do some action gerrit action.
 3. Check the directory for saved parameters.
 4. Execute locally the binary
    1. You need to set `GERRIT_HOOK_ARGFILE` environment variable to point to the saved file
    2. You need `/home/gerrit/.config/gerrit-hook/config.yaml` in your local home

Optional: You can also create `/tmp/gerrit-hook-log` to collect standard zap log in a file.