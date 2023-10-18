#

``` text
.______    __  ___  __    __  .______
|   _  \  |  |/  / |  |  |  | |   _  \
|  |_)  | |  '  /  |  |  |  | |  |_)  |__ _  ___ _ __
|   ___/  |    <   |  |  |  | |   ___// _' |/ _ \ '_ \
|  |      |  .  \  |  '--'  | |  |   | (_| |  __/ | | |
| _|      |__|\__\  \______/  | _|    \__, |\___|_| |_|
                                      |___/
```

---

[![license](https://img.shields.io/badge/License-MIT-brightgreen.svg?style=for-the-badge)](https://github.com/pPrecel/pkup-gen/blob/main/LICENSE)
[![build](https://img.shields.io/github/actions/workflow/status/pPrecel/pkup-gen/tests-build.yml?style=for-the-badge)](https://github.com/pPrecel/pkup-gen/actions/workflows/build.yml)

---

Simple and easy-to-use tool to generate PKUP (`Podwyższone Koszty Uzyskania Przychodu` - Polish law thing) artifacts, `.patch` files, based on merged Github PullRequests.

The `pkup-gen` collects all users' PullRequests merged between the 18th (23:59:59) of the actual month and the 19th (00:00:00) of the past one. To qualify PR, the user should be an author or committer of at least one commit from the PullRequest.

## Installation

Visit the [releases page](https://github.com/pPrecel/pkup-gen/releases) to download one of the pre-built binaries for your platform.

### Homebrew

1. Install the `pkup-gen` using the Homebrew:

    ```bash
    brew install pPrecel/tap/pkup-gen
    ```

    or

    ```bash
    brew tap pPrecel/tap
    brew install pkup-gen
    ```

## Usage

Example usage:

```text
pkup gen --token "<PAT_TOKEN>" --username "pPrecel" \
    --repo "kyma-project/serverless-manager" \
    --repo "kyma-project/keda-manager" \
    --repo "kyma-project/warden" \
    --repo "kyma-incubator/reconciler" \
    --repo "kyma-project/test-infra" \
    --repo "kyma-project/kyma" \
    --with-closed
.______    __  ___  __    __  .______
|   _  \  |  |/  / |  |  |  | |   _  \
|  |_)  | |  '  /  |  |  |  | |  |_)  |__ _  ___ _ __
|   ___/  |    <   |  |  |  | |   ___// _' |/ _ \ '_ \
|  |      |  .  \  |  '--'  | |  |   | (_| |  __/ | | |
| _|      |__|\__\  \______/  | _|    \__, |\___|_| |_|
                                      |___/     v1.2.0

INFO  generating artifacts for the actual PKUP period
    ├ after: 2023-09-19 00:00:00
    └ before: 2023-10-18 23:59:59
 ✓  found 11 PRs for repo 'kyma-project/serverless-manager'
      ├──󰘭 (#351) Reflect used presets in status
      ├──󰘭 (#348) Fix function default preset
      ├──󰘭 (#313) Get rid of setup
      ├──󰘭 (#318) Remove unnecessary ifs from the `delete.go` file
      ├──󰘭 (#317) Make loggers more consistent
      ├──󰘭 (#312) Improve building flags mechanism
      ├──󰘭 (#304) Improve optional dependencies state function
      ├──󰘭 (#288) Apply linter suggestions
      ├──󰘭 (#291) Rename the `stopWithError` func
      ├──󰘭 (#287) Use requeueAfter secret are deleted
      └──󰘭 (#244) Implement module-config template
 ✓  found 2 PRs for repo 'kyma-project/keda-manager'
      ├──󰘭 (#299) Add more release logs
      └──󰘭 (#285) Implement module-config template
 ✓  found 7 PRs for repo 'kyma-project/warden'
      ├──󰘭 (#119) Add unit tests for the `certs` package
      ├──󰘭 (#117) Refactor webhook secret strategy
      ├──󰘭 (#118) Add possibility to export cover out file
      ├──󰘭 (#113) Sec scanners config
      ├── (#121) Bugfixes 0.5
      ├── (#114) Warden module poc
      └── (#115) Warden module poc 2
 ✓  found 2 PRs for repo 'kyma-project/test-infra'
      ├──󰘭 (#9046) Fix `warden-unit-test` job
      └──󰘭 (#9035) Add missing argument to the warden dind job
 ✗  skipping 'kyma-project/kyma' no user activity detected
 ✗  skipping 'kyma-incubator/reconciler' no user activity detected
INFO  all patch files saved to dir
    └ dir: /Users/pprecel/go/src/github.com/pPrecel/pkup-gen
```

output:

```bash
ls --tree

.
├── kyma-project_keda-manager.patch
├── kyma-project_serverless-manager.patch
├── kyma-project_test-infra.patch
└── kyma-project_warden.patch
```

## Personal Access Token

The `pkup-gen` is using GitHub API for all HTTP operations. It does mean that to generate artifacts you have to pass a [PAT](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens) using the `--token` flag. For public projects, the generated token does not need to have any permissions.

If the token is not specified the `pkup-gen` will try to connect your GitHub account with the [pkup-gen](https://github.com/apps/pkup-gen) app using the GitHub device API. The generated token will be saved on your local machine so next time, until the token expires, you will be logged in. Example run without the `--token` flag specified:

```text
pkup gen --username "pPrecel" \
    --repo "kyma-project/warden" --with-closed
.______    __  ___  __    __  .______
|   _  \  |  |/  / |  |  |  | |   _  \
|  |_)  | |  '  /  |  |  |  | |  |_)  |__ _  ___ _ __
|   ___/  |    <   |  |  |  | |   ___// _' |/ _ \ '_ \
|  |      |  .  \  |  '--'  | |  |   | (_| |  __/ | | |
| _|      |__|\__\  \______/  | _|    \__, |\___|_| |_|
                                      |___/      v1.2.0

WARN  no token provided - grand access via pkup-gen GitHub app
    ├ copy code: A16E-C1UI
    └ then open and paste the above code: https://github.com/login/device
INFO  generating artifacts for the actual PKUP period
    ├ after: 2023-09-19 00:00:00
    └ before: 2023-10-18 23:59:59
 ✓  found 7 PRs for repo 'kyma-project/warden'
      ├──󰘭 (#119) Add unit tests for the `certs` package
      ├──󰘭 (#117) Refactor webhook secret strategy
      ├──󰘭 (#118) Add possibility to export cover out file
      ├──󰘭 (#113) Sec scanners config
      ├── (#121) Bugfixes 0.5
      ├── (#114) Warden module poc
      └── (#115) Warden module poc 2
INFO  all patch files saved to dir
    └ dir: /Users/pprecel/go/src/github.com/pPrecel/pkup-gen
```
