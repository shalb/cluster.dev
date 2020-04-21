# Enable Github auth in ArgoCD

## Adding Auth Provider
_details https://argoproj.github.io/argo-cd/operator-manual/user-management/#sso_

Edit ArgoCD configmap and set the `dex.config`:
```bash 
kubectl edit configmap argocd-cm -n argocd
```
`clientID` and `clientSecret` should be obtained during creation of a Github Oauth App:
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
_details: https://argoproj.github.io/argo-cd/operator-manual/rbac/_

After login you'll receive authentication with login ex: `voa@shalb.com`, with your Github group setting, ex:`shalb:dev`
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

