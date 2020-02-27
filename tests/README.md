# Testing locally

## Run actions local for tests

1. Create `config.sh` and setup required variables

```bash
cp config.example.sh config.sh
```

2. Run tests

```bash
./tests.sh
```

## Destroy infrastructure, created by terraform

```bash
cd ../terraform/aws/(module) && terraform destroy
# (then press enter for leave variables empty, set only region)
```
