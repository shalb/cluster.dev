FROM hashicorp/terraform:full
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]