FROM hashicorp/terraform:light as terraform

### Install Helmfile
# Image pulled from https://hub.docker.com/r/chatwork/helmfile/dockerfile
# TODO create own image with terraform and helmfile versioning

FROM chatwork/helmfile:latest

COPY --from=terraform /bin/terraform /bin/terraform

ENV PRJ_ROOT /app
WORKDIR $PRJ_ROOT
# Look on .dockerignore file to check what included
COPY . .

ENTRYPOINT ["/app/entrypoint.sh"]
