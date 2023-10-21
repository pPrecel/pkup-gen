# Simple Flow

This scenario shows how to generate `.diff` files for a few repositories for the actual PKUP period in the acctual directory.

1. Construct `pkup-gen` spell with most basic flags:

    ```bash
    pkup gen --username "<GITHUB_USERNAME>" --repo "<ORG>/<REPO>"
    ```

    > Example: pkup gen --username "pPrecel" --repo "kyma-project/serverless-manager"

2. Because PAT token is not provided, `pkup-gen` try to connect to the GitHub app to create one - here follow the instructions in the `WARN` log:

    ![1](../../assets/screenshot-simple-flow-1.png)

3. On GitHub side you have to pass copied code and click `Continue` and then `Authorize pkup-gen`:

    ![2](../../assets/screenshot-simple-flow-2.png)
    ![3](../../assets/screenshot-simple-flow-3.png)

4. Now `pkup-gen` continues to generate artifacts what is represented in logs:

    ![4](../../assets/screenshot-simple-flow-4.png)

5. At the end you can see a few new files generated in the acctual directory and the `.txt` file with all info to manually fill `.docx` report:

    ![5](../../assets/screenshot-simple-flow-5.png)
    ![6](../../assets/screenshot-simple-flow-6.png)
