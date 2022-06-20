# Secrets

Secret is an object that contains sensitive data such as a password, a token, or a key. It is used to pass secret values to the tools that don't have a proper support of secret engines.

Cluster.dev allows for two ways of working with secrets.  

## SOPS secrets

Cluster.dev uses SOPS binary to **create** and **edit** SOPS secrets. However, the SOPS binary is **not required** to use and decrypt SOPS secrets.

!!! note
     Since none of Cluster.dev reconciliation processes (build, plan, apply) requires SOPS, you don't have to install it for the pipelines.

See [SOPS installation instructions](https://github.com/mozilla/sops#download) in official repo.

Secrets are encoded/decoded with [SOPS](https://github.com/mozilla/sops) utility that supports AWS KMS, GCP KMS, Azure Key Vault and PGP keys. How to use:

1. Use Cluster.dev console client to create a new secret from scratch:

     ```bash
     cdev secret create
     ```

2. Use interactive menu to create a secret.

3. Edit the secret and set secret data in `encrypted_data:` section.

4. Use references to the secret data in a stack template (you can find the examples in the generated secret file).

## AWS Secrets Manager

Cluster.dev client can use AWS SSM as a secret storage. How to use:

1. Create a new secret in AWS Secrets Manager using AWS CLI or web console. Both raw and JSON data formats are supported.

2. Use Cluster.dev console client to create a new secret from scratch:

     ```bash
     cdev secret create
     ```

3. Answer the questions. For the `Name of secret in AWS Secrets Manager` enter the name of the AWS secret created above.

4. Use references to the secret data in a stack template (you can find the examples in the generated secret file).

To list and edit any secret, use the commands:

```bash
cdev secret ls
```

and

```bash
cdev secret edit secret_name
```

## Secrets reference

You can refer to a secret data in stack files with {{ .secrets.secret_name.secret_key }} syntax.   

For example, we have a secret in AWS Secrets Manager and want to refer to the secret in our `stack.yaml`:

```yaml
name: my-aws-secret
kind: Secret
driver: aws_secretmanager
spec: 
    region: eu-central-1
    aws_secret_name: pass
```

In order to do this, we need to define the secret as {{ .secrets.my-aws-secret.some-key }} in the `stack.yaml`:

```yaml
name: my-stack
template: https://<template.git.url>
kind: Stack
variables:
  region: eu-central-1
  name: my-test-stack
  password: {{ .secrets.my-aws-secret.some-key }}
```



