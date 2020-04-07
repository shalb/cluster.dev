# Installer

Implement [cli-installer specification](../../docs/design/cli-installer-design.md)

**Status: `In progress`**

## Style guide

Code style checked by pylint

Pylint settings:

```json
    "python.linting.pylintArgs": [
        "--disable=E0401"
    ],
```

Function documentation style based on [google style example](https://sphinxcontrib-napoleon.readthedocs.io/en/latest/example_google.html).

## Local debug and run

### Fist build

```bash
docker-compose build
```

### Run

```bash
docker-compose run app
```

## Useful links

* [PyInquirer (interactive part) examples](https://github.com/CITGuru/PyInquirer#examples)
* [GitPython Tutorial (for creation repo for user)](https://gitpython.readthedocs.io/en/stable/tutorial.html)
* [Creating standalone Mac OS X applications](https://www.metachris.com/2015/11/create-standalone-mac-os-x-applications-with-python-and-py2app/) | [py2app docs](https://py2app.readthedocs.io/en/latest/)
* [py2exe - Create standalone Windows app](https://www.py2exe.org/)
