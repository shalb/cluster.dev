FROM python:3.8.2-alpine3.11

WORKDIR /app
RUN apk add --no-cache \
        git \
        openssh \
        # For PyInquirer
        gcc \
        musl-dev

# .ssh dir mounts with 1000:1000 rights. And can't be accessible with other rights.
RUN addgroup -g 1000 cluster.dev && \
    adduser -u 1000 -D -G cluster.dev cluster.dev
USER cluster.dev

COPY requirements.txt .
RUN pip install -r requirements.txt

COPY cluster_dev.py .
ENTRYPOINT ["python", "cluster_dev.py"]
CMD ["-h"]
