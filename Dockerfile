ARG HELMFILE_VERSION=0.113.0
ARG GO_VERSION=1.16.0-alpine
ARG TERRAFORM_VERSION=0.14.7

FROM golang:${GO_VERSION} as builder

RUN apk add --no-cache make git bash
WORKDIR /workspace/cluster-dev
COPY . /workspace/cluster-dev
RUN make

FROM hashicorp/terraform:${TERRAFORM_VERSION} as terraform

COPY --from=builder /workspace/cluster-dev/bin/cdev /bin/cdev

WORKDIR /workspace/cluster-dev

ENTRYPOINT ["/bin/cdev"]
