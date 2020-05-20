FROM python:3.8.2-alpine3.11

WORKDIR /app
RUN apk add --no-cache \
        git \
        openssh \
        # For PyInquirer
        gcc \
        musl-dev \
        # GPG for sign commits
        gnupg \
        # For PyNaCl
        libffi-dev \
        make \
        # Editor
        neovim && \
        ln -s /usr/bin/nvim /usr/bin/editor

COPY requirements.txt .
RUN pip --no-cache-dir install -r requirements.txt

COPY entrypoint.sh cluster_dev.py aws_policy.json ./

ENTRYPOINT ["/app/entrypoint.sh"]
CMD [ "-h" ]
