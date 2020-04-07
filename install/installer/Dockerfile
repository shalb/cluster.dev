FROM python:3.8.2-alpine3.11

WORKDIR /app
RUN apk add --no-cache \
        git \
        # For PyInquirer
        gcc \
        musl-dev

COPY requirements.txt .
RUN pip install -r requirements.txt

COPY cluster_dev.py .
ENTRYPOINT ["python", "cluster_dev.py"]
CMD ["-h"]
