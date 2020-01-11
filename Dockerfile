FROM hashicorp/terraform:light
RUN apk add --no-cache git bash openssh

# Install aws-cli
RUN apk -Uuv add groff less python py-pip
RUN pip install awscli
RUN apk --purge -v del py-pip
RUN rm /var/cache/apk/*

COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]