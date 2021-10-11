# Cdev Install Reference

## Download from release

Each stable version of cdev has a binary that can be downloaded and installed manually. The documentation is suitable for **v0.4.0 or higher** of cluster.dev client.

Installation example for Linux amd64:

1. Download your desired version from the [releases page](https://github.com/shalb/cluster.dev/releases).

2. Unpack it.

3. Find the cdev binary in the unpacked directory.

4. Move the cdev binary to bin folder (/usr/local/bin/).

## Building from source

Go version 16 or higher is required, see [Golang installation instructions](https://golang.org/doc/install).

To build cluster.dev client from source:

1. Clone cluster.dev Git repo:

     ```bash
     git clone https://github.com/shalb/cluster.dev/
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
