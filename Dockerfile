FROM hashicorp/terraform:light as terraform

### Install Helmfile
# Image pulled from https://hub.docker.com/r/chatwork/helmfile/dockerfile
# TODO create own image with terraform and helmfile versioning

FROM ubuntu:18.04
ARG YAMLTOENV_VERSION=v0.0.3
ARG DEBIAN_FRONTEND=noninteractive
COPY --from=terraform /bin/terraform /bin/terraform

### Install so
RUN apt-get -y update && apt-get -y install curl s3cmd git awscli \
    && curl -L  https://github.com/shalb/yamltoenv/releases/download/${YAMLTOENV_VERSION}/yamltoenv_${YAMLTOENV_VERSION}_linux_amd64.tgz |  tar -xvz \
    && rm -f yamltoenv_${YAMLTOENV_VERSION}_linux_amd64.tgz && mv yamltoenv /bin/ \
    && curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl \
    && chmod +x ./kubectl && mv ./kubectl /usr/local/bin/kubectl

ENV PRJ_ROOT /app
WORKDIR $PRJ_ROOT

# Look on .dockerignore file to check what included
COPY . .

ENTRYPOINT ["/app/entrypoint.sh"]
