ARG HELMFILE_VERSION=0.113.0

FROM golang:1.14.2-alpine3.11 as builder

RUN apk add --no-cache make git bash
WORKDIR /workspace/cluster-dev
COPY . /workspace/cluster-dev
RUN make

FROM hashicorp/terraform:0.12.29 as terraform

### Install Helmfile
# Image pulled from https://hub.docker.com/r/chatwork/helmfile/dockerfile
# TODO create own image with terraform and helmfile versioning

FROM chatwork/helmfile:$HELMFILE_VERSION

COPY --from=terraform /bin/terraform /bin/terraform
COPY --from=builder /workspace/cluster-dev/bin/reconciler /bin/reconciler

### Install s3cmd
RUN python3 -m pip install --upgrade pip && \
    pip3 install --no-cache-dir --upgrade s3cmd
ENV PRJ_ROOT /app
WORKDIR $PRJ_ROOT

# Look on .dockerignore file to check what included
COPY . .

ENTRYPOINT ["/bin/reconciler"]
