# Testing locally

## Run actions local for tests

1. Create `config.sh` and setup required variables.

```bash
cp config.example.sh config.sh
```

2. Run tests

```bash
./tests.sh
```

It create cluster infrastructure that specified in your github workflow file (set in step 1).

## Destroy infrastructure, created by terraform

When you end, destroy all resources created during test.

```bash
terraform destroy ../terraform/aws/argocd
# (then press enter for leave variables empty, set only region)
terraform destroy ../terraform/aws/minikube
# (then press enter for leave variables empty, set only region)
terraform destroy ../terraform/aws/backend
# (then press enter for leave variables empty, set only region)
```

**Atention**: This is example script, please, recheck that you remove all resources.
