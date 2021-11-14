# Secrets

Secret is an object that contains sensitive data such as a password, a token, or a key. Is used to pass secret values to the tools that don't have a proper support of secret engines.

There are two ways to use secrets:

## SOPS secrets

For **creating** and **editing** SOPS secrets, Cluster.dev uses SOPS binary. But the SOPS binary is **not required** for decrypting and using SOPS secrets. As none of Cluster.dev reconcilation processes (build, plan, apply) requires SOPS to be performed, you don't have to install it for pipelines.

See [SOPS installation instructions](https://github.com/mozilla/sops#download) in official repo.

Secrets are encoded/decoded with [SOPS](https://github.com/mozilla/sops) utility that supports AWS KMS, GCP KMS, Azure Key Vault and PGP keys. How to use:

1. Use Cluster.dev console client to create a new secret from scratch:

     ```bash
     cdev secret create
     ```

2. Use interactive menu to create a secret.

3. Edit the secret and set secret data in `encrypted_data:` section.

4. Use references to the secret data in a stack template (you can find the examples in the generated secret file).

## Amazon secret manager

Cluster.dev client can use AWS SSM as a secret storage. How to use:

1. Create a new secret in AWS secret manager using AWS CLI or web console. Both raw and JSON data formats are supported.

2. Use Cluster.dev console client to create a new secret from scratch:

     ```bash
     cdev secret create
     ```

3. Answer the questions. For `Name of secret in AWS Secrets manager` enter the name of the AWS secret created above.

4. Use references to the secret data in a stack template (you can find the examples in the generated secret file).

To list and edit any secret, use the commands:

```bash
cdev secret ls
```

and

```bash
cdev secret edit secret_name
```
