FROM cli-installer

USER root
RUN apk add --no-cache py3-pytest
USER cluster.dev

COPY test_installer.py .

ENTRYPOINT ["pytest"]
