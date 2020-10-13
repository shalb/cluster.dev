#!/usr/bin/env python3.8
# flake8: noqa  # FIXME - this disable all checks for file!
"""cluster.dev installation script.

Example usage:
    ./cluster.dev.py install
    ./cluster.dev.py install -h
    ./cluster.dev.py install -p Github
"""
import argparse
import glob
import json
import os
import shutil
import sys
from contextlib import suppress
from pathlib import Path
from typing import TYPE_CHECKING

import aws_config
import github
import validate
import wget
from aws import create_user_and_permissions as aws_create_user_and_permissions
from git import GitCommandError
from git import InvalidGitRepositoryError
from git import Repo
from logger import logger
from validate import ask_user

try:
    from typeguard import typechecked  # noqa: WPS433
except ModuleNotFoundError:
    def typechecked(func=None):  # noqa: WPS440
        """Skip runtime type checking on the function arguments."""
        return func

if TYPE_CHECKING or typechecked:
    from git.cmd import Git  # noqa: WPS433


#######################################################################
#                          F U N C T I O N S                          #
#######################################################################


@typechecked
def dir_is_git_repo(path: str) -> bool:
    """Check if directory is a Git repository.

    Args:
        path: (str) path to directory.

    Returns:
        bool: True if successful, False otherwise.
    """
    try:
        Repo(path)
    except InvalidGitRepositoryError:
        return False

    return True


@typechecked
def parse_cli_args() -> argparse.Namespace:  # noqa: WPS213 # TODO
    """Parse CLI arguments, validate it.

    Returns:
        argparse.Namespace: object that contains all params entered by user.
    """
    parser = argparse.ArgumentParser(
        usage=''
        + '  interactive:     ./cluster-dev.py install\n'
        + '  non-interactive: ./cluster-dev.py install -gp Github',
    )

    parser.add_argument('subcommand', nargs='+', metavar='install', choices=['install'])

    parser.add_argument(
        '--git-provider',
        metavar='<provider>',
        dest='git_provider',
        help='Can be Github, Bitbucket or Gitlab',
        choices=GIT_PROVIDERS,
    )
    parser.add_argument(
        '--create-repo',
        metavar='<repo_name>',
        dest='repo_name',
        help='Automatically initialize repo for you',
    )
    parser.add_argument(
        '--git-user-name',
        metavar='<username>',
        dest='git_user_name',
        help='Username used in Git Provider.' +
        'Can be automatically get from .gitconfig',
    )
    parser.add_argument(
        '--git-password',
        metavar='<password>',
        dest='git_password',
        help='Password used in Git Provider. ' +
        'Can be automatically get from .ssh',
    )
    parser.add_argument(
        '--git-token',
        metavar='<token>',
        dest='git_token',
        help='Token for API access. For ex. GITHUB_TOKEN',
    )
    parser.add_argument(
        '--cloud',
        metavar='<cloud>',
        dest='cloud',
        help='Can be AWS or DigitalOcean',
        choices=CLOUDS,
    )
    parser.add_argument(
        '--cloud-provider',
        metavar='<provider>',
        dest='cloud_provider',
        help='Cloud provider depends on selected --cloud',
    )
    parser.add_argument(
        '--cloud-programatic-login',
        metavar='<ACCESS_KEY_ID>',
        dest='cloud_login',
        default='',
        help='AWS_ACCESS_KEY_ID or SPACES_ACCESS_KEY_ID',
    )
    parser.add_argument(
        '--cloud-programatic-password',
        metavar='<SECRET_ACCESS_KEY>',
        dest='cloud_password',
        default='',
        help='AWS_SECRET_ACCESS_KEY or SPACES_SECRET_ACCESS_KEY',
    )
    parser.add_argument(
        '--cloud-token',
        metavar='<TOKEN>',
        dest='cloud_token',
        help='SESSION_TOKEN or DIGITALOCEAN_TOKEN',
    )
    parser.add_argument(
        '--cloud-user',
        metavar='<user>',
        dest='cloud_user',
        help='User name which be created/used for cluster.dev. Applicable only for AWS. '
        + 'If specified user exist, --cloud-programatic-login and --cloud-programatic-password '
        + 'will be try use as it AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.',
    )
    parser.add_argument(
        '--cluster-dev-version',
        metavar='<cluster.dev release>',
        dest='release_version',
        help='Specify cluster.dev release or last realease will be used',
    )

    cli = parser.parse_args()

    if cli.repo_name:
        validate.RepoName.validate(cli.repo_name, interactive=False)

    if cli.git_user_name:
        validate.UserName.validate(cli.git_user_name, interactive=False)

    if cli.cloud and cli.cloud_provider:
        if cli.cloud_provider not in CLOUD_PROVIDERS[cli.cloud]:
            # TODO: refactor https://github.com/shalb/cluster.dev/pull/70#discussion_r431407132
            sys.exit(
                f'Cloud provider can be: {CLOUD_PROVIDERS[cli.cloud_provider]}, '
                + f'but provided: {cli.cloud_provider}',
            )

    return cli


@typechecked
def get_git_username(git: Git) -> str:
    """Get username from settings or from user input.

    Args:
        git: (git.cmd.Git) Manages communication with the Git binary.

    Returns:
        username string.
    """
    try:  # Required mount: $HOME/.gitconfig:/home/cluster.dev/.gitconfig:ro
        user = git.config('--get', 'user.name')
    except GitCommandError:
        user = ask_user(
            name='git_user_name',
            type='input',
            message='Please enter your git username',
            validate=validate.UserName,
        )
    else:
        logger.info(f'Username: {user}')

    return user


@typechecked
def get_git_password() -> str:
    """Get SSH key from settings or password from user input.

    Returns:
        empty string if SSH-key provided (mounted).
        Otherwise - return password string.
    """
    with suppress(FileNotFoundError):
        # Try use ssh as password
        # Required mount: $HOME/.ssh:/home/cluster.dev/.ssh:ro
        os.listdir(f'{os.environ["HOME"]}/.ssh')
        logger.info('Password type: ssh-key')
        return ''

    return ask_user(
        name='git_password',
        type='password',
        message='Please enter your git password',
    )


@typechecked
def remove_all_except_git(dir_path: str):  # noqa: WPS231 # FIXME
    """Remove all in directory except `.git/`.

    Args:
        dir_path: (str) path to dir which will cleanup.
    """
    for filename in os.listdir(dir_path):
        path = os.path.join(dir_path, filename)
        try:
            if os.path.isfile(path) or os.path.islink(path):
                os.unlink(path)

            elif os.path.isdir(path):
                if path == os.path.join(dir_path, '.git'):
                    continue

                shutil.rmtree(path)

        except Exception as exception:  # pylint: disable=broad-except
            logger.warning(f'Failed to delete {path}. Reason: {exception}')


@typechecked
def choose_git_provider(repo: Repo) -> str:
    """Choose git provider.

    Try automatically, if git remotes specified. Otherwise - ask user.

    Args:
        repo: (git.Repo) Package with general repository related functions

    Returns:
        git_provider string.
    """
    if not repo.remotes:  # Remotes not exist in locally init repos
        return ask_user(
            name='choose_git_provider',
            type='list',
            message='Select your Git Provider',
            choices=GIT_PROVIDERS,
        )

    git = repo.git
    remote = git.remote('-v')

    if remote.find('github'):
        git_provider = 'Github'
    elif remote.find('bitbucket'):
        git_provider = 'Bitbucket'
    elif remote.find('gitlab'):
        git_provider = 'Gitlab'
    else:
        git_provider = ask_user(
            name='choose_git_provider',
            type='list',
            message='Select your Git Provider',
            # choices=GIT_PROVIDERS,  # noqa: E800 #? Should be use after implementation
            choices=[
                {
                    'name': 'Github',
                }, {
                    'name': 'Bitbucket',
                    'disabled': 'Unavailable at this time',
                }, {
                    'name': 'Gitlab',
                    'disabled': 'Unavailable at this time',
                },
            ],
        )

    return git_provider


@typechecked
def get_git_token(provider: str) -> str:
    """Get github token from user input or settings.

    Args:
        provider: (str) Provider name.

    Returns:
        git_token string.
    """
    # Required env variables from host: GITHUB_TOKEN
    if provider == 'Github':
        git_token = os.environ.get('GITHUB_TOKEN')
        if git_token is not None:
            logger.info('Use GITHUB_TOKEN from env')
        else:
            git_token = ask_user(
                name='github_token',
                type='password',
                message='Please enter GITHUB_TOKEN. '
                + 'It can be generated at https://github.com/settings/tokens',
            )

    if not git_token:
        sys.exit(f'ERROR: Provider "{provider}" not exist in function "get_git_token"')

    return git_token


@typechecked
def get_repo_name_from_url(url: str) -> str:
    """Get git repository name from git remote url.

    Args:
        url: (str) git remote url. Can be "git@" or "https://".

    Returns:
        str: git repository name.
    """
    # TODO: Refactor https://github.com/shalb/cluster.dev/pull/70#discussion_r427309906
    last_slash_index = url.rfind('/')
    last_suffix_index = url.rfind('.git')

    if last_slash_index < 0 or last_suffix_index < 0:
        sys.exit(f'ERROR: Check `git remote -v`. Badly formatted origin: {url}')

    return url[last_slash_index + 1:last_suffix_index]


@typechecked
def get_repo_owner_from_url(url: str) -> str:
    """Get git repository owner from git remote url.

    Args:
        url: (str) git remote url. Can be "git@" or "https://".

    Returns:
        str: git repository owner (user or organization).
    """
    last_slash_index = url.rfind('/')

    # IF HTTPS: https://github.com/shalb/cluster.dev.git
    prefix_index = url.find('/', 8)
    # IF SSH: git@github.com:shalb/cluster.dev.git
    if prefix_index == last_slash_index:
        prefix_index = url.find(':')

    if last_slash_index < 0 or prefix_index < 0:
        sys.exit(f'ERROR: Check `git remote -v`. Badly formatted origin: {url}')

    return url[prefix_index + 1:last_slash_index]


@typechecked
def add_sample_cluster_dev_files(  # noqa: WPS210, WPS211 # FIXME
    cloud: str,
    cloud_provider: str,
    git_provider: str,
    repo_path: str,
    release: str,
):
    """Populate repo with sample files.

    Args:
        cloud: (str) Cloud name.
        cloud_provider: (str) Cloud provider name.
        git_provider: (str) Git Provider name.
        repo_path: (str) Relative path to repo.
        release: (str) cluster.dev release version. Can be tag commit or branch.
    """
    repo_url = f'https://raw.githubusercontent.com/shalb/cluster.dev/{release}'

    # Create CI config dir and add files
    if git_provider == 'Github':
        ci_path = os.path.join(repo_path, '.github', 'workflows')
        Path(ci_path).mkdir(parents=True, exist_ok=True)

        ci_file_url = f'{repo_url}/.github/workflows/aws.yml'

    ci_file_path = os.path.join(ci_path, 'main.yml')
    wget.download(ci_file_url, ci_file_path)
    # User should have rights to edit file
    # But get inside Docker real uif:gid into Docker not trivial
    os.chmod(ci_file_path, 0o666)  # noqa: WPS432, S103

    # Create cluster.dev config dir and add files
    config_path = os.path.join(repo_path, '.cluster.dev')
    Path(config_path).mkdir(parents=True, exist_ok=True)

    if cloud == 'AWS':
        if cloud_provider == 'Minikube':
            config_file_name = 'aws-minikube.yaml'

    config_file_url = f'{repo_url}/.cluster.dev/{config_file_name}'

    config_file_path = os.path.join(config_path, config_file_name)
    wget.download(config_file_url, config_file_path)
    # User should have rights to edit file
    # But get inside Docker real uif:gid into Docker not trivial
    os.chmod(config_file_path, 0o666)  # noqa: WPS432, S103

    logger.info('\nLocal repo populated with sample files')


@typechecked
def commit_and_push(git: Git, message: str):
    """Stage, commit and push all changes.

    Args:
        git: (git.cmd.Git) Manages communication with the Git binary.
        message: (str) Commit message body.
    """
    git.add('-A')
    try:
        git.commit('-m', message)
    except GitCommandError:
        logger.warning('Nothing to commit, working tree clean')

    git.push()

    logger.info(message)


@typechecked
def last_edited_config_path(repo_path: str) -> str:
    """Get path to last edited .cluster.dev config.

    Args:
        repo_path: (str) Relative path to repo.

    Returns:
        str: Relative path to last edited config.
    """
    # *.yaml means all files with .yaml extention
    config_mask = os.path.join(repo_path, '.cluster.dev', '*.yaml')
    config_files = glob.glob(config_mask)

    return max(config_files, key=os.path.getmtime)


@typechecked
def set_cluster_installed(config_path: str, installed: bool):
    """Set cluster.installed option in cluster config file.

    Args:
        config_path: (str) Relative path to .cluster.dev config.
        installed: (bool) Value that should be set to `installed` option.
    """
    # Change `installed` value
    prev_value = json.dumps(not installed)
    new_value = json.dumps(installed)

    with open(config_path, 'r') as cfo:
        conf = cfo.read()

    new_conf = conf.replace(f'installed: {prev_value}', f'installed: {new_value}')

    with open(config_path, 'w') as cfn:
        cfn.write(new_conf)


@typechecked
def main() -> None:  # pylint: disable=too-many-statements,R0914 # noqa: WPS231, WPS213, WPS210
    """Logic."""
    cli = parse_cli_args()
    dir_path = os.path.relpath('current_dir')

    logger.info('Hi, we gonna create an infrastructure for you.\n')

    if not dir_is_git_repo(dir_path):
        create_repo = True
        if not cli.repo_name:
            logger.info('As this is a GitOps approach we need to start with the git repo.')
            create_repo = ask_user(
                type='list',
                message='Create repo for you?',
                choices=[
                    {
                        'name': 'No, I create or clone repo and then run tool there',
                        'value': False,
                    }, {
                        'name': 'Yes',
                        'value': True,
                        'disabled': 'Unavailable at this time',
                    },
                ],
            )

        if not create_repo:
            sys.exit('OK. See you soon!')

        repo_name = cli.repo_name or ask_user(
            type='input',
            message='Please enter the name of your infrastructure repository',
            default='infrastructure',
            validate=validate.RepoName,
        )

        # TODO: setup remote origin and so on. Can be useful:
        # user = cli.git_user_name or get_git_username(git)  # noqa: E800
        # password = cli.git_password or get_git_password()  # noqa: E800
        sys.exit('TODO')

    logger.info('Inside git repo, use it.')

    repo = Repo(dir_path)
    git = repo.git

    if repo.heads:  # Heads exist only after first commit
        cleanup_repo = ask_user(
            type='confirm',
            message='This is not empty repo. Delete all existing configurations?',
            default=False,
        )

        if cleanup_repo:
            remove_all_except_git(dir_path)
            git.add('-A')
            git.commit('-m', 'Cleanup repo')

    git_provider = cli.git_provider or choose_git_provider(repo)
    git_token = cli.git_token or get_git_token(git_provider)

    if not repo.remotes:
        publish_repo = ask_user(
            name='publish_repo_to_git_provider',
            type='confirm',
            message='Your repo not published to Git Provider yet. Publish it now?',
            default=True,
        )
        if publish_repo:
            # TODO: push repo to Git Provider
            sys.exit('TODO')

    user = cli.git_user_name or get_git_username(git)  # pylint: disable=W0612 # noqa: F841 # TODO
    password = cli.git_password or get_git_password()  # pylint: disable=W0612 # noqa: F841 # TODO

    cloud = cli.cloud or ask_user(
        type='list',
        message='Select your Cloud',
        # choices=CLOUDS  # noqa: E800 #? Should be use after implementation
        choices=[
            {
                'name': 'AWS',
            }, {
                'name': 'DigitalOcean',
                'disabled': 'Unavailable at this time',
            },
        ],
    )
    cloud_user = cli.cloud_user or ask_user(
        name='aws_cloud_user',
        type='input',
        message='Please enter username for cluster.dev user',
        validate=validate.AWSUserName,
        default='cluster.dev',
    )

    if cloud == 'AWS':
        config = aws_config.parse_file(cli.cloud_login, cli.cloud_password)
        config_section = aws_config.choose_section(config)

        access_key = cli.cloud_login or aws_config.get_login(config, config_section)
        secret_key = cli.cloud_password or aws_config.get_password(config, config_section)
        session_token = (
            cli.cloud_token
            or aws_config.get_session(config, config_section, cli.cloud_login)
        )

        release_version = cli.release_version or github.get_last_release()

        creds = aws_create_user_and_permissions(
            cloud_user, access_key, secret_key, session_token, release_version,
        )
        if creds['created']:
            logger.info(
                f'Credentials for user "{cloud_user}":\n'
                + f'aws_access_key_id={creds["key"]}\n'
                + f'aws_secret_access_key={creds["secret"]}',
            )
    # elif cloud == 'DigitalOcean':  # noqa: E800
        # TODO
        # https://www.digitalocean.com/docs/apis-clis/doctl/how-to/install/
        # cloud_login = cli.cloud_login or get_do_login()  # noqa: E800
        # cloud_password = cli.cloud_password or get_do_password()  # noqa: E800
        # cloud_token = cli.cloud_token or ask_user(  # noqa: E800
        #     type='password',  # noqa: E800
        #     message='Please enter your Cloud token',  # noqa: E800
        # )  # noqa: E800

    cloud_provider = (
        cli.cloud_provider
        or ask_user(
            type='list',
            message='Select your Cloud Provider',
            choices=CLOUD_PROVIDERS[cloud],
        )
    )

    remote = repo.remotes.origin.url
    owner = get_repo_owner_from_url(remote)
    repo_name = get_repo_name_from_url(remote)

    if git_provider == 'Github':
        github.create_secrets(creds, cloud, owner, repo_name, git_token)

    add_sample_cluster_dev_files(cloud, cloud_provider, git_provider, dir_path, release_version)
    commit_and_push(git, 'cluster.dev: Add sample files')

    config_path = last_edited_config_path(dir_path)
    set_cluster_installed(config_path, installed=True)
    # Open editor
    os.system(f'editor "{config_path}"')  # noqa: S605
    commit_and_push(git, 'cluster.dev: Up cluster')

    # Show link to logs. Temporary solution
    if git_provider == 'Github':
        logger.info(f'See logs at: https://github.com/{owner}/{repo_name}/actions')


#######################################################################
#                         G L O B A L   A R G S                       #
#######################################################################
GIT_PROVIDERS = ('Github')  # , 'Bitbucket', 'Gitlab')
CLOUDS = ('AWS')  # , 'DigitalOcean')

CLOUD_PROVIDERS = {  # noqa: WPS407
    'AWS': (
        'Minikube',
        'AWS EKS',
    ),
    'DigitalOcean': (
        'Minikube',
        'Managed Kubernetes',
    ),
}

if __name__ == '__main__':
    # execute only if run as a script
    main()
