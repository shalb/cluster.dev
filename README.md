# Cluster.dev - Cloud Native Infrastructure Orchestration

<!-- markdownlint-disable no-inline-html-->
<p align="center">
    <img src="https://raw.githubusercontent.com/shalb/cluster.dev/master/docs/images/cluster-dev-logo-site.png?sanitize=true"
        width="540">
</p>
<p align="center">
    <a href="https://join.slack.com/t/cluster-dev/shared_invite/zt-eg4q6jae-v0~zgrBLYTTXt~CjnjmprA" alt="Join Slack">
        <img src="https://img.shields.io/static/v1?label=SLACK&message=JOIN&color=4A154B&style=for-the-badge&logo=slack" /></a>
    <a href="https://twitter.com/intent/follow?screen_name=shalbcom">
        <img src="https://img.shields.io/static/v1?label=TWITTER&message=FOLLOW&color=1DA1F2&style=for-the-badge&logo=twitter"
            alt="follow on Twitter"></a>
    <a href="https://www.facebook.com/shalb/">
        <img src="https://img.shields.io/static/v1?label=FACEBOOK&message=FOLLOW&color=1877F2&style=for-the-badge&logo=facebook"
            alt="follow on Facebook"></a>
</p>

Cluster.dev is an open-source tool used to easily create complete dev/stage/prod environments from infrastructure templates. The templates could be based on Terraform modules, Kubernetes manifests, Shell scripts, Helm, Kustomize and ArgoCD apps.

We provide glue that could stick those components together, so you can deploy, test and distribute a whole set of components with pinned versions.

----

## MENU <!-- omit in toc -->

* [How does it work?](#how-does-it-work)
* [How does it differ?](#how-does-it-differ)
  * [How does it differ from Pulumi?](#how-does-it-differ-from-pulumi)
  * [How does it differ from Terraform?](#how-does-it-differ-from-terraform)
* [Code of Conduct and License](#code-of-conduct-and-license)

----

## How does it work?

You can create or download a predefined template,    Then set the variables,        then render and deploy a whole infra set    
<image template sample>                            +   <infra variables>        =      <image with infra in cloud AWS/K8s>

Ok, what else can it do?

0. Re-use all existing Terraform private and public modules and Helm Charts.
 
1. We can apply parallel changes in multiple infrastructures concurrently:
<image with parallel reconciliation>
 
2. We can use the same global variables and secrets across different infrastructures, clouds and technologies.
<image with variables in dev/stage>    <image with same variable in terraform/kubernetes>   <image same secrets in AWS/GCP>       
                 
3. Template anything with Go-template function, even Terraform modules in Helm style templates
<sample invocation of terraform module templated with helm functions>
 
4. Render GitOps configuration for applications and push it to the repos connected to a target infrastructure.

## How does it differ?

### How does it differ from Pulumi?

To work with Pulumi you need the knowledge of NodeJs, Go or .NET. With cluster.dev it is enough to know Terraform, Kubernetes or Shell scripts. You can even do without it at all - just use ready templates to create what you need and where you need.

### How does it differ from Terraform?

## Code of Conduct and License

Code of Conduct is described in [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md).

Product is licensed under [Apache 2.0](./LICENSE).
