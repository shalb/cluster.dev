# Cluster.dev - Cloud Infrastructures' Management Tool

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

Cluster.dev is an open-source tool designed to manage cloud native infrastructures with simple declarative manifests - infrastructure templates. The infrastructure templates could be based on Terraform modules, Kubernetes manifests, Shell scripts, Helm charts, Kustomize and ArgoCD/Flux applications, OPA policies etc. Cluster.dev sticks those components together so that you could deploy, test and distribute a whole set of components with pinned versions.

For more information on cluster.dev and its functionality please visit [docs.cluster.dev](https://docs.cluster.dev/). 

### When do I need Cluster.dev?

1. If you have a common infrastructure pattern that contains multiple components stuck together.
   Like a bunch of TF-modules, or a set of K8s addons. So you need to re-use this pattern inside your projects.
2. If you develop an infrastructure platform that you ship to other teams, and they need to launch new infras from your template.
3. If you build a complex infrastructure that contains different technologies, and you need to perform integration testing to confirm the components' interoperability. After which you can promote the changes to next environments.
4. If you are a software vendor and you need to deliver infrastructure deployment along with your software.

### Principle Diagram

![cdev diagram](./docs/images/cdev-base-diagram.png)
