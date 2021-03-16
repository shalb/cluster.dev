# Installing cdev

## Installing cdev with a single command

```bash
export CDEV_VERSION=$(curl -s https://api.github.com/repos/shalb/cluster.dev/releases/latest | grep tag_name | cut -d '"' -f 4) && curl -Lo cdev.tgz https://github.com/shalb/cluster.dev/releases/download/${CDEV_VERSION}/cluster.dev_${CDEV_VERSION}_linux_amd64.tgz && tar -xzvf cdev.tgz -C /usr/local/bin
```

## Downloading from release

Binaries of the latest stable release are available on the [releases page](https://github.com/shalb/cluster.dev/releases). The documentation is suitable for **v0.4.0 or higher** of cluster.dev client.

Installation example for Linux amd64:

```bash
wget https://github.com/shalb/cluster.dev/releases/download/v0.4.0-rc1/cluster.dev_v0.4.0-rc1_linux_amd64.tgz
tar -xzvf cluster.dev_v0.4.0-rc1_linux_amd64.tgz -C /usr/local/bin

cdev --help
```

## Building from source

Go version 16 or higher is required. [Golang installation instructions](https://golang.org/doc/install).

To build cluster.dev client from source:

1. Clone cluster.dev git repo:

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
