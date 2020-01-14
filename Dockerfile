FROM hashicorp/terraform:light as terraform

### Install Helmfile
# code pulled from https://hub.docker.com/r/chatwork/helmfile/dockerfile

FROM chatwork/helmfile:latest

COPY --from=terraform /usr/local/bin/terraform /usr/local/bin/terraform

COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]