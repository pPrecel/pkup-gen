# Compose and Send

This scenario describes how to generate many reports for given repositories and organizations and send them to developers.

> NOTE: knowledge from the [simple-scenario](../simple-flow/README.md) and the [with-template](../with-template/README.md) may be useful to clarify some topics.

1. Create the `.pkupgencompose.yaml` file with the whole configuration ( config details [here](../../pkg/config/config.go) ):

    ```yaml
    reports:
    - outputDir: reports/FILIP_STROZIK
      email: "filip.strozik@outlook.com"
      signatures:
      - username: pPrecel
      - username: internalPrecel
        enterpriseUrl: "https://github.my-corp"
      reportFields:
        pkupGenEmployeesName: "Filip Strózik"
        pkupGenJobTitle: "Senior Developer"
        pkupGenDepartment: "R&D"
        pkupGenManagersName: "John Wick"
    # - outputDir: ...
    
    template: templates/report.docx

    orgs:
    - name: kyma-project
      token: ghp_5...C
    - name: kyma-incubator
      token: ghp_5...C
    - name: kyma
      token: ghp_1...G
      enterpriseUrl: "https://github.my-corp"
    
    repos:
    - name: kyma-project/busola
      token: ghp_5...C
      allBranches: true
      uniqueOnly: true
    - name: kyma-project/cli
      token: ghp_5...C
      branches: ["main", "v3"]
      uniqueOnly: true
    
    send:
      serverAddress: "smtp-mail.outlook.com"
      serverPort: 587
      username: "filip.strozik@outlook.com"
      password: *************
      subject: "PKUP report"
      htmlBodyPath: "templates/email_body_template.html"
      from: filip.strozik@outlook.com
    ```

2. Compose report ( example output ):

    ![1](../../assets/screenshot-compose-and-send-1.png)

3. Send email ( example output ):

    ![2](../../assets/screenshot-compose-and-send-2.png)
