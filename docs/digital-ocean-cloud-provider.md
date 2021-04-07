# Deploying to DigitalOcean

## Authentication

Create [an access token](https://www.digitalocean.com/docs/apis-clis/api/create-personal-access-token/).

For details on using DO spaces bucket as a backend, see [here](https://www.digitalocean.com/community/questions/spaces-as-terraform-backend).

## Set up access to DigitalOcean

1) Install `doctl`. For details see [here](https://www.digitalocean.com/docs/apis-clis/doctl/how-to/install/).

    ```bash
    cd ~
    wget https://github.com/digitalocean/doctl/releases/download/v1.57.0/doctl-1.57.0-linux-amd64.tar.gz
    tar xf ~/doctl-1.57.0-linux-amd64.tar.gz
    sudo mv ~/doctl /usr/local/bin
    ```

2) Clone the repo and change directory to example dir:

    ```bash
    git clone https://github.com/shalb/cluster.dev.git
    cd cluster.dev/examples/do_k8s/
    ```

3) Export your DIGITALOCEAN_TOKEN, for details see [here](https://www.digitalocean.com/docs/apis-clis/api/create-personal-access-token/).

    ```bash
    export DIGITALOCEAN_TOKEN="MyDIGITALOCEANToken" 
    ```

4) Export SPACES_ACCESS_KEY_ID and SPACES_SECRET_ACCESS_KEY environment variables, for details see [here](https://www.digitalocean.com/community/tutorials/how-to-create-a-digitalocean-space-and-api-key).

    ```bash
    export SPACES_SECRET_ACCESS_KEY="dSUGdbJqa6xwJ6Fo8qV2DSksdjh..."
    export SPACES_SECRET_ACCESS_KEY="TEaKjdj8DSaJl7EnOdsa..."
    ```

5) Create a spaces bucket for Terraform states in the chosen region (in the example we used the 'cdev-data' bucket name). (https://www.digitalocean.com/docs/spaces/quickstart/#create-a-space)

6) Create a domain in DigitalOcean domains service (in the example we used 'k8s.cluster.dev' zone, you can change it). (https://www.digitalocean.com/docs/networking/dns/how-to/add-domains/)

7) Use `cdev` to deploy infrastructure:

    ```bash
    cdev plan
    cdev apply
    ```
