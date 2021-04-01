# Installing cdev

## From script

cdev has an installer script that takes the latest version of cdev and installs it for you locally. You can fetch the script and execute it locally:

```bash
curl -fsSL https://raw.githubusercontent.com/shalb/cluster.dev/master/scripts/get_cdev.sh | bash
```

## Download from release

Each stable version of cdev has a binary that can be downloaded and installed manually. The documentation is suitable for **v0.4.0 or higher** of cluster.dev client.

Installation example for Linux amd64:

1. Download your desired version from the [releases page](https://github.com/shalb/cluster.dev/releases) (as an example we'll have v0.4.0):

    ```bash
    export CDEV_VERSION=v0.4.0-rc1
    ```

2. Unpack it:

    ```bash
    tar -xzvf cluster.dev_${CDEV_VERSION}_linux_amd64
    ```

3. Move the cdev binary to its desired destination:

    ```bash
    sudo mv cdev /usr/local/bin/
    ```

## Building from source

Go version 16 or higher is required, see [Golang installation instructions](https://golang.org/doc/install).

To build cluster.dev client from source:

1. Clone cluster.dev Git repo:

     ```bash
     git clone --depth 1 --branch v0.4.0-rc1 https://github.com/shalb/cluster.dev/
     ```

2. Build the binary:

     ```bash
     cd cluster.dev/ && make
     ```

3. Check cdev and move the binary to bin folder:

     ```bash
     ./bin/cdev --help
     mv ./bin/cdev /usr/local/bin/
     ```
