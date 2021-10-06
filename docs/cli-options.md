# CLI Options

## Global flags

* `--cache`             Use previously cached build directory.

* `-l, --log-level string`   Set the logging level ('debug'|'info'|'warn'|'error'|'fatal') (default "info").

* `--parallelism int`    Max parallel threads for module applying (default - `3`).

* `--trace`              Print functions trace info in logs (mainly used for development).

## Apply flags

* `--force`              Skip interactive approval.

* `-h`, `--help`         Help for apply.

* `--ignore-state`       Apply even if the state has not changed.

## Create flags

* `-h`, `--help`        Help for create.

* `--interactive`       Use interactive mode for project generation.

* `--list-templates`    Show all available stack templates for project generator.

## Destroy flags

* `--force`              Skip interactive approval.

* `-h`, `--help`         Help for destroy.

* `--ignore-state`       Destroy current configuration and ignore state.
