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
    --repo "kyma-incubator/reconciler" \
    --repo "kyma-project/test-infra" \
    --repo "kyma-project/kyma"

.______    __  ___  __    __  .______
|   _  \  |  |/  / |  |  |  | |   _  \
|  |_)  | |  '  /  |  |  |  | |  |_)  |__ _  ___ _ __
|   ___/  |    <   |  |  |  | |   ___// _' |/ _ \ '_ \
|  |      |  .  \  |  '--'  | |  |   | (_| |  __/ | | |
| _|      |__|\__\  \______/  | _|    \__, |\___|_| |_|
                                      |___/

INFO  generating artifacts for the actual PKUP period
    ├ after: 2023-09-19 00:00:00
    └ before: 2023-10-18 23:59:59
 ✓  found 10 PRs for repo 'kyma-project/serverless-manager'
      ├──Fix function default preset
      ├──Get rid of setup
      ├──Remove unnecessary ifs from the `delete.go` file
      ├──Make loggers more consistent
      ├──Improve building flags mechanism
      ├──Improve optional dependencies state function
      ├──Apply linter suggestions
      ├──Rename the `stopWithError` func
      ├──Use requeueAfter secret are deleted
      └──Implement module-config template
 ✓  found 2 PRs for repo 'kyma-project/keda-manager'
      ├──Add more release logs
      └──Implement module-config template
 ✓  found 2 PRs for repo 'kyma-project/test-infra'
      ├──Fix `warden-unit-test` job
      └──Add missing argument to the warden dind job
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
└── kyma-project_test-infra.patch
```

## Personal Access Token

The `pkup-gen` is using GitHub API for all HTTP operations. It does mean that to generate artifacts you have to pass a [PAT](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens) using the `--token` flag. For public projects, the generated token does not need to have any permissions.
