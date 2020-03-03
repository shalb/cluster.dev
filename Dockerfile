FROM hashicorp/terraform:light as terraform

### Install Helmfile
# Image pulled from https://hub.docker.com/r/chatwork/helmfile/dockerfile
# TODO create own image with terraform and helmfile versioning

FROM chatwork/helmfile:latest

COPY --from=terraform /bin/terraform /bin/terraform

COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
