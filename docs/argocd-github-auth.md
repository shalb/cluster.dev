## Adding Auth Provider

For details please see the [Argo CD documentation](https://argoproj.github.io/argo-cd/operator-manual/user-management/#sso_).

Edit ArgoCD configmap and set the `dex.config`:
```bash
kubectl edit configmap argocd-cm -n argocd
```
`clientID` and `clientSecret` should be obtained while creating a Github Oauth App:

```yaml
  dex.config: |
    connectors:
      - type: github
        id: github
        name: GitHub
        config:
          clientID: 00000000000000000000
          clientSecret: 000000000000000000000000000000000000
          orgs:
          - name: shalb
```

## Mapping Group Permissions
For details please see the [Argo CD documentation](https://argoproj.github.io/argo-cd/operator-manual/rbac/).

After login you'll receive authentication with login, ex: `voa@shalb.com`, with your GitHub group setting, ex:`shalb:dev`
So you can define its permission in ArgoCD project manifest:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: AppProject
spec:
  clusterResourceWhitelist:
  - group: '*'
    kind: '*'
  destinations:
  - namespace: '*'
    server: '*'
  sourceRepos:
  - '*'
  roles:
  - description: Read-only privileges to default
    groups:
    - shalb:dev
    name: read-only
    policies:
    - p, proj:default:read-only, applications, get, default/*, allow

```
