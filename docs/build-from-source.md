# Build from source

Go version 16 or higher is required. [Golang installation instructions](https://golang.org/doc/install).

To build cluster.dev client from source:

Clone cluster.dev git repo:

```bash
git clone --depth 1 --branch v0.4.0-rc1 https://github.com/shalb/cluster.dev/
```

Build binary:

```bash
cd cluster.dev/ && make
```

Check cdev and move binary to bin folder:

```bash
./bin/cdev --help
mv ./bin/cdev /usr/local/bin/
```
