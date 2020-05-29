#!/usr/bin/env python3
"""cluster.dev installation script.

Example usage:
    ./cluster.dev.py install
    ./cluster.dev.py install -h
    ./cluster.dev.py install -p Github
"""
import argparse
import glob
import json
import logging
import os
import shutil
import sys
from configparser import ConfigParser
from configparser import NoOptionError
from contextlib import suppress
from pathlib import Path
from typing import Dict
from typing import TYPE_CHECKING
from typing import Union

import boto3
import github
import validate
import wget
from git import GitCommandError
from git import InvalidGitRepositoryError
from git import Repo
from PyInquirer import prompt
from PyInquirer import style_from_dict
from PyInquirer import Token
from typeguard import typechecked

if TYPE_CHECKING or typechecked:
    from git.cmd import Git  # noqa: WPS433

logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)
logger.addHandler(logging.StreamHandler(sys.stdout))

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
def ask_user(question_name: str, choices=None):
    """Draw menu for user interactions.

    Args:
        question_name: (str) Name from `questions` that will processed.
        choices: Used when choose params become know during execution.

    Returns:
        Depends on entered `question_name`:
            create_repo: (bool) True or False.
            git_provider (string): Provider name. See `GIT_PROVIDERS` for variants.
            repo_name (string,int): Repo name. Have built-in validator,
            so user can't enter invalid repository name.
    """
    prompt_style = style_from_dict({
        Token.Separator: '#6C6C6C',
        Token.QuestionMark: '#FF9D00 bold',
        # Token.Selected: '',  # default
        Token.Selected: '#5F819D',
        Token.Pointer: '#FF9D00 bold',
        Token.Instruction: '',  # default
        Token.Answer: '#5F819D bold',
        Token.Question: '',
    })

    # Interactive prompt questions
    # See https://github.com/CITGuru/PyInquirer#examples for usage variants
    #
    # 'NAME' should be same.
    # Example:
    #    `'NAME': {
    #        'name': 'NAME'
    #    }`
    #
    questions = {
        'create_repo': {
            'type': 'list',
            'name': 'create_repo',
            'message': 'Create repo for you?',
            'choices': [
                {
                    'name': 'No, I create or clone repo and then run tool there',
                    'value': False,
                }, {
                    'name': 'Yes',
                    'value': True,
                    'disabled': 'Unavailable at this time',
                },
            ],
        },
        'choose_git_provider': {
            'type': 'list',
            'name': 'choose_git_provider',
            'message': 'Select your Git Provider',
            'choices': GIT_PROVIDERS,
        },
        'repo_name': {
            'type': 'input',
            'name': 'repo_name',
            'message': 'Please enter the name of your infrastructure repository',
            'default': 'infrastructure',
            'validate': validate.RepoName,
        },
        'cleanup_repo': {
            'type': 'confirm',
            'name': 'cleanup_repo',
            'message': 'This is not empty repo. Delete all existing configurations?',
            'default': False,
        },
        'publish_repo_to_git_provider': {
            'type': 'confirm',
            'name': 'publish_repo_to_git_provider',
            'message': 'Your repo not published to Git Provider yet. Publish it now?',
            'default': True,
        },
        'git_user_name': {
            'type': 'input',
            'name': 'git_user_name',
            'message': 'Please enter your git username',
            'validate': validate.UserName,
        },
        'git_password': {
            'type': 'password',
            'name': 'git_password',
            'message': 'Please enter your git password',
        },
        'choose_cloud': {
            'type': 'list',
            'name': 'choose_cloud',
            'message': 'Select your Cloud',
            'choices': CLOUDS,
        },
        'choose_cloud_provider': {
            'type': 'list',
            'name': 'choose_cloud_provider',
            'message': 'Select your Cloud Provider',
            'choices': choices,
        },
        'cloud_login': {  # TODO: Add validation (and for cli arg)
            'type': 'input',
            'name': 'cloud_login',
            'message': 'Please enter your Cloud programatic key',
        },
        'cloud_password': {  # TODO: Add validation (and for cli arg)
            'type': 'password',
            'name': 'cloud_password',
            'message': 'Please enter your Cloud programatic secret',
        },
        'cloud_token': {
            'type': 'password',
            'name': 'cloud_token',
            'message': 'Please enter your Cloud token',
        },
        'choose_config_section': {
            'type': 'list',
            'name': 'choose_config_section',
            'message': 'Select credentials section that will be used to deploy cluster.dev',
            'choices': choices,
        },
        'aws_session_token': {
            'type': 'password',
            'name': 'aws_session_token',
            'message': 'Please enter your AWS Session token',
        },
        'aws_cloud_user': {
            'type': 'input',
            'name': 'aws_cloud_user',
            'message': 'Please enter username for cluster.dev user',
            'validate': validate.AWSUserName,
            'default': 'cluster.dev',
        },
        'aws_secret_key': {
            'type': 'password',
            'name': 'aws_secret_key',
            'message': f'Please enter AWS Secret Key for {choices}',
        },
        'github_token': {
            'type': 'password',
            'name': 'github_token',
            'message': 'Please enter GITHUB_TOKEN. '
            + 'It can be generated at https://github.com/settings/tokens',
        },
    }

    try:
        answer = prompt(questions[question_name], style=prompt_style)[question_name]
    except KeyError:
        sys.exit(f"Sorry, it's program error. Can't found key '{question_name}'")

    return answer


@typechecked
def parse_cli_args() -> argparse.Namespace:
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
        user = ask_user('git_user_name')
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

    return ask_user('git_password')


@typechecked
def remove_all_except_git(dir_path: str):
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
        return ask_user('choose_git_provider')

    git = repo.git
    remote = git.remote('-v')

    if remote.find('github'):
        git_provider = 'Github'
    elif remote.find('bitbucket'):
        git_provider = 'Bitbucket'
    elif remote.find('gitlab'):
        git_provider = 'Gitlab'
    else:
        git_provider = ask_user('choose_git_provider')

    return git_provider


@typechecked
def get_data_from_aws_config(login: str, password: str) -> {bool, ConfigParser}:
    """Get and parse config from file.

    Args:
        login: (str) CLI argument provided by user.
        password: (str) CLI argument provided by user.

    Returns:
        config configparser.ConfigParser|False: config, if exists. Otherwise - False.
    """
    # Skip if CLI args provided
    if login and password:
        return False

    path = f'{os.environ["HOME"]}/.aws/credentials'
    try:  # Required mount: $HOME/.aws/:/home/cluster.dev/.aws/:ro
        os.path.isfile(path)
    except FileNotFoundError:
        return False

    config = ConfigParser()
    config.read(path)

    return config


@typechecked
def get_aws_config_section(config: {ConfigParser, bool}) -> str:
    """Ask user which section in config should be use for extracting credentials.

    Args:
        config (configparser.ConfigParser|False): INI config.

    Returns:
        str: section name, if exists. Otherwise - empty string.
    """
    # Skip if CLI args provided or configs doesn't exist
    if not config or not config.sections():
        return ''

    if len(config.sections()) == 1:
        section = config.sections()[0]
        logger.info(f'Use AWS creds from file, section "{section}"')
    else:
        # User can have multiply creds, ask him what should be used
        # https://boto3.amazonaws.com/v1/documentation/api/latest/guide/configuration.html#configuring-credentials
        section = ask_user('choose_config_section', choices=config.sections())

    return section


@typechecked
def get_aws_login(config: {ConfigParser, bool}, config_section: str) -> str:
    """Get cloud programatic login from settings or from user input.

    Args:
        config (configparser.ConfigParser|False): INI config.
        config_section: (str) INI config section.

    Returns:
        cloud_login string.
    """
    if not config or not config_section:
        return ask_user('cloud_login')

    return config.get(config_section, 'aws_access_key_id')


@typechecked
def get_aws_password(config: {ConfigParser, bool}, config_section: str) -> str:
    """Get cloud programatic password from settings or from user input.

    Args:
        config: (configparser.ConfigParser|False) INI config.
        config_section: (str) INI config section.

    Returns:
        cloud_password string.
    """
    if not config:
        return ask_user('cloud_password')

    return config.get(config_section, 'aws_secret_access_key')


@typechecked
def get_aws_session(
    config: {ConfigParser, bool},
    config_section: str,
    mfa_disabled: str = '',
) -> str:
    """Get cloud session from settings or from user input.

    Args:
        config: (configparser.ConfigParser|False) INI config.
        config_section: (str) INI config section.
        mfa_disabled: (str) CLI argument provided by user.
            If not provided - set to empty string.

    Returns:
        session_token string.
    """
    # If login provided but session - not, user may not have MFA enabled
    if mfa_disabled:
        logger.info('SESSION_TOKEN not found, try without MFA')
        return ''

    if not config:
        return ask_user('aws_session_token')

    try:
        session_token = config.get(config_section, 'aws_session_token')
    except NoOptionError:
        logger.info('SESSION_TOKEN not found, try without MFA')
        return ''

    return session_token


@typechecked
def create_aws_user_and_permissions(
    user: str,
    login: str,
    password: str,
    session: str,
    release: str,
) -> Dict[str, Union[str, bool]]:
    """Create cloud user and attach needed permissions.

    Args:
        user: (str) Cluster.dev user name.
        login: (str) Cloud programatic login.
        password: (str) Cloud programatic password.
        session: (str) Cloud session token.
        release: (str) cluster.dev release tag.

    Returns:
        Dict[str, Union[str, bool]]:
        `{'key': 'AWS_ACCESS_KEY_ID', 'secret': 'AWS_SECRET_ACCESS_KEY', 'created': bool}`.
    """
    iam = boto3.client(
        'iam',
        aws_access_key_id=login,
        aws_secret_access_key=password,
        aws_session_token=session,
    )
    keys_created = False
    link = (
        f'https://github.com/shalb/cluster.dev/blob/{release}'
        + '/install/installer_aws_install_req_permissions.json'
    )

    # If specified user exist, --cloud-programatic-login and
    # --cloud-programatic-password will be try use as it

    # AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
    try:  # Required "iam:ListUsers"???, "iam:ListAccessKeys"
        response = iam.list_access_keys(UserName=user)
    except iam.exceptions.NoSuchEntityException:  # User not yet exist
        logger.debug(f'User "{user}" does not exist yet')
    except iam.exceptions.ClientError as error:
        sys.exit(f'\nERROR: {error}\n\nRights what you need described here: \n{link}')
    else:
        try:
            if response['AccessKeyMetadata'][0]['AccessKeyId'] == login:
                logger.info(f'Specified creds belong to exist user "{user}", use it')

                return {
                    'key': login,
                    'secret': password,
                    'created': keys_created,
                }
        except IndexError:  # User have no programattic access keys
            # Required "iam:CreateAccessKey"
            response = iam.create_access_key(UserName=user)

            return {
                'key': response['AccessKey']['AccessKeyId'],
                'secret': response['AccessKey']['SecretAccessKey'],
                'created': True,
            }

    # Create/get policy_arn for user
    with open('aws_policy.json', 'r') as policy_file:
        policy = json.dumps(json.load(policy_file))

    try:  # Required "iam:CreatePolicy"
        response = iam.create_policy(
            PolicyName='cluster.dev-policy',
            Path='/cluster.dev/',
            PolicyDocument=policy,
            Description='Policy for https://cluster.dev propertly work. '
            + 'Created by CLI instalator',
        )
    except iam.exceptions.EntityAlreadyExistsException:  # Policy already exist
        try:  # Required "iam:ListPolicies"
            response = iam.list_policies(
                Scope='Local',
                PathPrefix='/cluster.dev/',
            )
        except iam.exceptions.ClientError as error:  # noqa: WPS440
            sys.exit(f'\nERROR: {error}\n\nRights what you need described here: \n{link}')
        else:
            policy_arn = response['Policies'][0]['Arn']

    except iam.exceptions.ClientError as error:  # noqa: WPS440
        sys.exit(f'\nERROR: {error}\n\nRights what you need described here: \n{link}')
    else:
        policy_arn = response['Policy']['Arn']
        logger.info('Policy created')

    try:  # Required "iam:GetUser"
        # When cluster.dev user created, but provided creds for another user
        iam.get_user(UserName=user)
    except iam.exceptions.NoSuchEntityException:
        # Create user and access keys
        try:  # Required "iam:CreateUser"
            iam.create_user(
                Path='/cluster.dev/',
                UserName=user,
            )
        except iam.exceptions.ClientError as error:  # noqa: WPS440
            sys.exit(f'\nERROR: {error}\n\nRights what you need described here: \n{link}')

        try:  # Required "iam:AttachUserPolicy"
            iam.attach_user_policy(
                UserName=user,
                PolicyArn=policy_arn,
            )
        except iam.exceptions.ClientError as error:  # noqa: WPS440
            sys.exit(f'\nERROR: {error}\n\nRights what you need described here: \n{link}')
        else:
            logger.info('User created')

            response = iam.create_access_key(UserName=user)
            key = response['AccessKey']['AccessKeyId']
            secret = response['AccessKey']['SecretAccessKey']
            keys_created = True

    except iam.exceptions.ClientError as error:  # noqa: WPS440
        sys.exit(f'\nERROR: {error}\n\nRights what you need described here: \n{link}')

    else:
        response = iam.list_access_keys(UserName=user)
        key = response['AccessKeyMetadata'][0]['AccessKeyId']

        secret = ask_user('aws_secret_key', choices={'user': user, 'key': key})

    return {
        'key': key,
        'secret': secret,
        'created': keys_created,
    }


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
            git_token = ask_user('github_token')

    return git_token


@typechecked
def get_repo_name_from_url(url: str) -> str:
    """Get git repository name from git remote url.

    Args:
        url: (str) git remote url. Can be "git@" or "https://".

    Returns:
        str: git repository name.
    """
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
def add_sample_cluster_dev_files(
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


@typechecked  # noqa: WPS231, WPS213, WPS210
def main() -> None:
    """Logic."""
    cli = parse_cli_args()
    dir_path = os.path.relpath('current_dir')

    logger.info('Hi, we gonna create an infrastructure for you.\n')

    if not dir_is_git_repo(dir_path):
        create_repo = True
        if not cli.repo_name:
            logger.info('As this is a GitOps approach we need to start with the git repo.')
            create_repo = ask_user('create_repo')

        if not create_repo:
            sys.exit('OK. See you soon!')

        repo_name = cli.repo_name or ask_user('repo_name')

        # TODO: setup remote origin and so on. Can be useful:
        # user = cli.git_user_name or get_git_username(git)  # noqa: E800
        # password = cli.git_password or get_git_password()  # noqa: E800
        sys.exit('TODO')

    logger.info('Inside git repo, use it.')

    repo = Repo(dir_path)
    git = repo.git

    if repo.heads:  # Heads exist only after first commit
        cleanup_repo = ask_user('cleanup_repo')

        if cleanup_repo:
            remove_all_except_git(dir_path)
            git.add('-A')
            git.commit('-m', 'Cleanup repo')

    git_provider = cli.git_provider or choose_git_provider(repo)
    git_token = cli.git_token or get_git_token(git_provider)

    if not repo.remotes:
        publish_repo = ask_user('publish_repo_to_git_provider')
        if publish_repo:
            # TODO: push repo to Git Provider
            sys.exit('TODO')

    user = cli.git_user_name or get_git_username(git)
    password = cli.git_password or get_git_password()

    cloud = cli.cloud or ask_user('choose_cloud')
    cloud_user = cli.cloud_user or ask_user('aws_cloud_user')

    if cloud == 'AWS':
        config = get_data_from_aws_config(cli.cloud_login, cli.cloud_password)
        config_section = get_aws_config_section(config)

        access_key = cli.cloud_login or get_aws_login(config, config_section)
        secret_key = cli.cloud_password or get_aws_password(config, config_section)
        session_token = cli.cloud_token or get_aws_session(config, config_section, cli.cloud_login)

        release_version = cli.release_version or github.get_last_release()

        creds = create_aws_user_and_permissions(
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
        # cloud_token = cli.cloud_token or ask_user('cloud_token')  # noqa: E800

    cloud_provider = (
        cli.cloud_provider
        or ask_user('choose_cloud_provider', choices=CLOUD_PROVIDERS[cloud])
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
GIT_PROVIDERS = ('Github')  # , 'Bitbucket', 'Gitlab']
CLOUDS = ('AWS', 'DigitalOcean')

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
