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

![screen1](./docs/Screenshot%202023-10-19%20at%2000.14.36.png)

output:

```bash
ls --tree

.
├── kyma-project_keda-manager.patch
├── kyma-project_serverless-manager.patch
└── kyma-project_warden.patch
```

## Personal Access Token

The `pkup-gen` is using GitHub API for all HTTP operations. It does mean that to generate artifacts you have to pass a [PAT](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens) using the `--token` flag. For public projects, the generated token does not need to have any permissions.

If the token is not specified the `pkup-gen` will try to connect your GitHub account with the [pkup-gen](https://github.com/apps/pkup-gen) app using the GitHub device API. The generated token will be saved on your local machine so next time, until the token expires, you will be logged in. Example run without the `--token` flag specified:

![screen2](./docs/Screenshot%202023-10-19%20at%2000.16.05.png)
