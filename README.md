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

Simple and easy to use tool to generate PKUP artifacts (`.patch` files) based on merged Github PullRequests.

## Installation

Visit the [releases page](https://github.com/pPrecel/pkup-gen/releases) to download one of the pre-built binaries for your platform.

### Homebrew

1. Use Homebrew to install `pkup-gen`:

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
                      --repo "kyma-project/kyma"

.______    __  ___  __    __  .______
|   _  \  |  |/  / |  |  |  | |   _  \
|  |_)  | |  '  /  |  |  |  | |  |_)  |__ _  ___ _ __
|   ___/  |    <   |  |  |  | |   ___// _' |/ _ \ '_ \
|  |      |  .  \  |  '--'  | |  |   | (_| |  __/ | | |
| _|      |__|\__\  \______/  | _|    \__, |\___|_| |_|
                                      |___/

INFO[0000] looking for changes beteen 2023-05-19 23:59:59 and 2023-06-19 23:59:59
INFO[0000] processing 'kyma-project/serverless-manager' repo
INFO[0006]      user 'pPrecel' is an author of 'Regenerate manifest when manager is changed'
INFO[0006]      user 'pPrecel' is an author of 'Group cache interface arguments'
INFO[0006]      user 'pPrecel' is an author of 'Reimplement cache strategy'
INFO[0006]      user 'pPrecel' is an author of 'Implement secret cache'
INFO[0006]      user 'pPrecel' is an author of 'Fix orphan check'
INFO[0006]      user 'pPrecel' is an author of 'Fix finalizer deletion'
INFO[0006]      user 'pPrecel' is an author of 'Fix release name'
INFO[0006]      user 'pPrecel' is an author of 'implement condition `Deleted`'
INFO[0008] patch saved to file '/Users/pprecel/go/src/github.com/pPrecel/PKUP/kyma-project_serverless-manager.patch'
INFO[0008] processing 'kyma-project/keda-manager' repo
INFO[0024]      user 'pPrecel' is an author of 'Remove keda after integration test finish'
INFO[0024]      user 'pPrecel' is an author of 'Regenerate config'
INFO[0024]      user 'pPrecel' is an author of 'Cover case when Keda CR is duplicated '
INFO[0025] patch saved to file '/Users/pprecel/go/src/github.com/pPrecel/PKUP/kyma-project_keda-manager.patch'
INFO[0025] processing 'kyma-project/kyma' repo
INFO[0068] skipping 'kyma-project/kyma' no user activity detected
INFO[0068] processing 'kyma-incubator/reconciler' repo
INFO[0087] skipping 'kyma-incubator/reconciler' no user activity detected
```

output:

```bash
ls --tree

.
├── kyma-project_keda-manager.patch
└── kyma-project_serverless-manager.patch
```
