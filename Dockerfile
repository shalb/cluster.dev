ARG HELMFILE_VERSION=0.113.0

FROM hashicorp/terraform:light as terraform

### Install Helmfile
# Image pulled from https://hub.docker.com/r/chatwork/helmfile/dockerfile
# TODO create own image with terraform and helmfile versioning

FROM chatwork/helmfile:$HELMFILE_VERSION
ENV YAMLTOENV_VERSION=v0.0.3

COPY --from=terraform /bin/terraform /bin/terraform

### Install s3cmd
RUN /usr/bin/python3.8 -m pip install --upgrade pip
RUN pip3 install --no-cache-dir --upgrade s3cmd

RUN curl -L  https://github.com/shalb/yamltoenv/releases/download/${YAMLTOENV_VERSION}/yamltoenv_${YAMLTOENV_VERSION}_linux_amd64.tgz |  tar -xvz \
    && rm -f yamltoenv_${YAMLTOENV_VERSION}_linux_amd64.tgz && mv yamltoenv /bin/

ENV PRJ_ROOT /app
WORKDIR $PRJ_ROOT

# Look on .dockerignore file to check what included
COPY . .

ENTRYPOINT ["/app/entrypoint.sh"]
