# Simple Flow

This scenario shows how to generate `.diff` files for a few repositories for the actual PKUP period in the actual directory.

1. Construct the `pkup-gen` spell with the most basic flags:

    ```bash
    pkup gen --username "<GITHUB_USERNAME>" --repo "<ORG>/<REPO>"
    ```

    > Example: pkup gen --username "pPrecel" --repo "kyma-project/serverless-manager"

2. Because PAT is not provided, `pkup-gen` tries to connect to the GitHub app to create one - here follow the instructions in the `WARN` log:

    ![1](../../assets/screenshot-simple-flow-1.png)
    On the GitHub side, you have to pass the copied code, click `Continue` and then `Authorize pkup-gen`:
    ![2](../../assets/screenshot-simple-flow-2.png)
    ![3](../../assets/screenshot-simple-flow-3.png)

3. Now `pkup-gen` continues to generate artifacts that are represented in logs:

    ![4](../../assets/screenshot-simple-flow-4.png)

4. At the end you can see a few new files generated in the actual directory and the `.txt` file with all the info to manually fill the `.docx` report:

    ![5](../../assets/screenshot-simple-flow-5.png)
    ![6](../../assets/screenshot-simple-flow-6.png)
