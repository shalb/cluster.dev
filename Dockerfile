FROM hashicorp/terraform:light as terraform

### Install Helmfile
# Image pulled from https://hub.docker.com/r/chatwork/helmfile/dockerfile
# TODO create own image with terraform and helmfile versioning

FROM chatwork/helmfile:0.106.3

COPY --from=terraform /bin/terraform /bin/terraform

### Install s3cmd
RUN pip3 install --no-cache-dir --upgrade s3cmd

ENV PRJ_ROOT /app
WORKDIR $PRJ_ROOT
# Look on .dockerignore file to check what included
COPY . .

ENTRYPOINT ["/app/entrypoint.sh"]
