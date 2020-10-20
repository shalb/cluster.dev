# Run reconciler locally

For quick development-and-test loop you'd like run reconciler locally, not in CD.

For this, needed:

1. [Install Golang 1.14 or higher](https://golang.org/doc/install)
2. [Install Terraform 0.12](https://releases.hashicorp.com/terraform/)

3. Export your Cloud credentials

    ```bash
    export AWS_ACCESS_KEY_ID=ATIAAJSXDBUVOQ4JR
    export AWS_SECRET_ACCESS_KEY=SuperAwsSecret
    ```

4. Build reconciler

    ```bash
    cd cluster.dev/
    make
    ```

5. Run reconciler on needed cluster-dev manifest

    ```bash
    ./bin/reconciler --config .cluster.dev/aws-minikube.yaml --log-level debug
    ```
