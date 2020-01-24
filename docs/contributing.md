# Development

## How to contribute 

1. Create the issue you are going to address in [GH Issues](https://github.com/shalb/cluster.dev/issues).


2. Spawn new branch from master named with GH Issue you are going to address ex, if issue #3 name the branch: `feature/GH-3`. 


3. Create a new cluster manifest in `.cluster.dev/gh-3.yaml, setting the name with target issue:
```yaml
cluster:
  name: gh-3
  installed: true
  cloud: 
    provider: aws
    region: eu-central-1
    vpc: default
    domain: shalb.net
  provisioner:
    type: minikube
    instanceType: m5.large
```

4. Create a new workflow in `.github/workflows` named with the feature ex: `gh-3.yaml`.  
Set the required branch name in placeholders and define `cluster-config` with the value `branch-name.yaml`, equals the filename from previous step, example:
```yaml
on:
  push:
      branches:
        - feature/GH-3
jobs:
  deploy_cluster_job:
    runs-on: ubuntu-latest
    name: Deploy and Update K8s Clusters
    steps:
    - name: Checkout Repo
      uses: actions/checkout@v2
      with:
        ref: 'feature/GH-3'
    - name: Reconcile Clusters
      id: reconcile
      uses: shalb/cluster.dev@feature/GH-3
      with:
        cluster-config: './.cluster.dev/gh-3.yaml'
        cloud-user: ${{ secrets.aws_access_key_id }}
        cloud-pass: ${{ secrets.aws_secret_access_key }}
    - name: Get the execution status
      run: echo "The status ${{ steps.reconcile.outputs.status }}"
```

5. Commit and push both files with comment, ex: `GH-3 Initial Commit`

6. Check the [GH Actions](https://github.com/shalb/cluster.dev/actions) for the env build process in the logs. And check if the cluster is ready in target provider.

7. Add your changes to code and open a Pull Request, assigning it to `voatsap` or `MaxymVlasov` for review.

8. After successful review merge branch to master with comment `Resolve GH-3`.

9. Delete all resources (ec2 instances, elastic ip's, etc..) assotiated with the issue.