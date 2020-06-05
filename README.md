# Storj CI

This repository contains CI environment for running Jenkins jobs.

It contains:
* multiple custom linters,
* setup for tools,
* setup for database.

## Update Process
Updates to linters or Go version in any storj repo should originate here.

### Linters
1. Make PR to here (storj/ci) with whatever for versions
2. Make PRs to other repos and to fix any problems
   a. If the PR is on github, bonus points are awarded if the PR description in the other repo links back this PR
3. Merge those PRs.
4. After those PRs are merged, merge the this storj/ci repo

### Go version
1. Update storj/ci with new go version
2. Make PRs in all other repos to:
   a. Use new Go version
   b. Use the branch in storj/ci for `docker.build` around line 5 of their Jenkinsfile.
3. After those PRs pass (not merged), merge PR in storj/ci
4. Remove branch change in all of the other PRs, they should still pass
5. Merge those PRs
