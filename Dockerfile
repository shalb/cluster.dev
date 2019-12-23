FROM alpine:3.10
#FROM hashicorp/terraform:light
# Copies your code file from your action repository to the filesystem path `/` of the container
COPY entrypoint.sh /entrypoint.sh
#COPY bin/yaml.sh /yaml.sh
# Code file to execute when the docker container starts up (`entrypoint.sh`)
ENTRYPOINT ["/entrypoint.sh"]