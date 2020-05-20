#!/usr/bin/env python3
"""
cluster.dev installation script

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
from base64 import b64encode
from configparser import ConfigParser
from configparser import NoOptionError
from pathlib import Path
from sys import exit as _exit_

import boto3
import regex
import wget
from agithub.GitHub import GitHub
from git import GitCommandError
from git import InvalidGitRepositoryError
from git import Repo
from nacl import encoding
from nacl import public
from PyInquirer import prompt
from PyInquirer import style_from_dict
from PyInquirer import Token
from PyInquirer import ValidationError
from PyInquirer import Validator
from typeguard import typechecked

#######################################################################
#                          F U N C T I O N S                          #
#######################################################################


@typechecked
def dir_is_git_repo(path: str) -> bool:
    """
    Check if directory is a Git repository
    Args:
        path (str): path to directory
    Returns:
        bool: True if successful, False otherwise.
    """
    try:
        Repo(path)
        return True
    except InvalidGitRepositoryError:
        return False


@typechecked
class RepoNameValidator(Validator):  # pylint: disable=too-few-public-methods
    """Validate user input"""

    @typechecked
    def validate(self: Validator, document=None, interactive: bool = True):
        """
        Validate user input string to Github repo name restrictions

        Raises:
            ValidationError: If intput string not match regex.
                Show error message for user and get him change fix error
        """
        repo_name = self
        if interactive:
            repo_name = document.text

        error_message = 'Please enter a valid repository name. ' + \
            'It should have 1-100 only latin letters, ' + \
            'numbers, underscores and dashes'

        okay = regex.match('^[A-Za-z0-9_-]{1,100}$', repo_name)

        if not okay:
            if not interactive:
                _exit_(error_message)

            raise ValidationError(
                message=error_message,
                cursor_position=len(repo_name),
            )  # Move cursor to end


@typechecked
class UserNameValidator(Validator):  # pylint: disable=too-few-public-methods
    """Validate user input"""

    @typechecked
    def validate(self: Validator, document=None, interactive: bool = True):
        """
        Validate user input string to Github repo name restrictions

        Raises:
            ValidationError: If intput string not match regex.
                Show error message for user and get him change fix error
        """
        username = self
        if interactive:
            username = document.text

        error_message = 'Please enter a valid username. ' + \
            'It should have 1-39 only latin letters, numbers, ' + \
            'single hyphens and cannot begin or end with a hyphen'

        okay = regex.match('^(?!-)[A-Za-z0-9-]{1,39}$', username)
        not_ok = regex.match('^.*[-]{2,}.*$', username)
        not_ok2 = regex.match('^.*-$', username)

        if not_ok or not_ok2 or not okay:
            if not interactive:
                _exit_(error_message)

            raise ValidationError(
                message=error_message,
                cursor_position=len(username),
            )  # Move cursor to end


@typechecked
class AWSUserNameValidator(Validator):  # pylint: disable=too-few-public-methods
    """Validate user input"""

    @typechecked
    def validate(self: Validator, document=None, interactive: bool = True):
        """
        Validate user input string to AWS username restrictions

        Raises:
            ValidationError: If intput string not match regex.
                Show error message for user and get him change fix error
        """
        username = self
        if interactive:
            username = document.text

        error_message = 'Please enter a valid AWS username. ' + \
            'It should have 1-64 only latin letters, ' + \
            'numbers and symbols: _+=,.@-'

        okay = regex.match('^[A-Za-z0-9_+=,.@-]{1,64}$', username)

        if not okay:
            if not interactive:
                _exit_(error_message)

            raise ValidationError(
                message=error_message,
                cursor_position=len(username),
            )  # Move cursor to end


@typechecked
def ask_user(question_name: str, choices=None):
    """
    Draw menu for user interactions

    Args:
        question_name (str): Name from `questions` that will processed.
        choices (list): used when choose params become know during execution
    Returns:
        Depends on entered `question_name`:
            create_repo (bool): True or False
            git_provider (string): Provider name. See `GIT_PROVIDERS` for variants
            repo_name (string,int): Repo name. Have built-in validator,
                so user can't enter invalid repository name
    Raises:
        KeyError: If `question_name` not exist in `questions`
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
    #    'NAME': {
    #        'name': 'NAME'
    #    }
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
                }, ],
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
            'validate': RepoNameValidator,
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
            'validate': UserNameValidator,
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
        'cloud_login': {
            'type': 'input',
            'name': 'cloud_login',
            'message': 'Please enter your Cloud programatic key',
        },
        'cloud_password': {
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
            'validate': AWSUserNameValidator,
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
            'message': 'Please enter GITHUB_TOKEN. It can be generated at https://github.com/settings/tokens',
        },
    }

    try:
        result = prompt(questions[question_name], style=prompt_style)[question_name]
    except KeyError:
        _exit_(f'Sorry, it\'s program error. Can\'t found key "{question_name}"')

    return result


@typechecked
def parse_cli_args() -> object:
    """
    Parse CLI arguments, validate it
    """
    parser = argparse.ArgumentParser(
        usage='' +
        '  interactive:     ./cluster-dev.py install\n' +
        '  non-interactive: ./cluster-dev.py install -gp Github',
    )

    parser.add_argument(
        'subcommand', nargs='+', metavar='install',
        choices=['install'],
    )
    parser.add_argument(
        '--git-provider', '-gp', metavar='<provider>',
        dest='git_provider',
        help='Can be Github, Bitbucket or Gitlab',
        choices=GIT_PROVIDERS,
    )
    parser.add_argument(
        '--create-repo', metavar='<repo_name>',
        dest='repo_name',
        help='Automatically initialize repo for you',
    )
    parser.add_argument(
        '--git-user-name', '-gusr', metavar='<username>',
        dest='git_user_name',
        help='Username used in Git Provider.' +
        'Can be automatically get from .gitconfig',
    )
    parser.add_argument(
        '--git-password', '-gpwd', metavar='<password>',
        dest='git_password',
        help='Password used in Git Provider. ' +
        'Can be automatically get from .ssh',
    )
    parser.add_argument(
        '--git-token', '-gtoken', metavar='<token>',
        dest='git_token',
        help='Token for API access. For ex. GITHUB_TOKEN',
    )
    parser.add_argument(
        '--cloud', '-c', metavar='<cloud>',
        dest='cloud',
        help="Can be AWS or DigitalOcean",
        choices=CLOUDS,
    )
    parser.add_argument(
        '--cloud-provider', '-cp', metavar='<provider>',
        dest='cloud_provider',
        help='Cloud provider depends on selected --cloud',
    )
    parser.add_argument(
        '--cloud-programatic-login', '-clogin', metavar='<ACCESS_KEY_ID>',
        dest='cloud_login', default='',
        help='AWS_ACCESS_KEY_ID or SPACES_ACCESS_KEY_ID',
    )
    parser.add_argument(
        '--cloud-programatic-password', '-cpwd', metavar='<SECRET_ACCESS_KEY>',
        dest='cloud_password', default='',
        help='AWS_SECRET_ACCESS_KEY or SPACES_SECRET_ACCESS_KEY',
    )
    parser.add_argument(
        '--cloud-token', '-ctoken', metavar='<TOKEN>',
        dest='cloud_token',
        help='SESSION_TOKEN or DIGITALOCEAN_TOKEN',
    )
    parser.add_argument(
        '--cloud-user', '-cuser', metavar='<user>',
        dest='cloud_user',
        help='User name which be created/used for cluster.dev. Applicable only for AWS. ' +
        'If specified user exist, -clogin and -cpwd will be try use as it ' +
        'AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.',
    )
    parser.add_argument(
        '--cluster-dev-version', '-version', metavar='<cluster.dev release>',
        dest='release_version',
        help='Specify cluster.dev release or last realease will be used',
    )

    cli = parser.parse_args()

    if cli.repo_name:
        RepoNameValidator.validate(cli.repo_name, interactive=False)

    if cli.git_user_name:
        UserNameValidator.validate(cli.git_user_name, interactive=False)

    if cli.cloud and cli.cloud_provider:
        if cli.cloud_provider not in CLOUD_PROVIDERS[cli.cloud]:
            _exit_(
                f'Cloud provider can be: {CLOUD_PROVIDERS[cli.cloud_provider]}, ' +
                f'but provided: {cli.cloud_provider}',
            )

    return cli


@typechecked
def get_git_username(git: object) -> str:
    """
    Get username from settings or from user input

    Args:
        git (obj): git.Repo.git object
    Returns:
        username string
    """
    try:
        # Required mount: $HOME/.gitconfig:/home/cluster.dev/.gitconfig:ro
        user = git.config('--get', 'user.name')
        print(f'Username: {user}')
    except GitCommandError:
        user = ask_user('git_user_name')

    return user


@typechecked
def get_git_password() -> str:
    """
    Get SSH key from settings or password from user input

    Returns:
        empty string if SSH-key provided (mounted)
        Otherwise - return password string
    """
    try:
        # Try use ssh as password
        # Required mount: $HOME/.ssh:/home/cluster.dev/.ssh:ro
        if os.listdir(f'{os.environ["HOME"]}/.ssh'):
            print('Password type: ssh-key')
            return ''

    except FileNotFoundError:
        pass

    password = ask_user('git_password')
    return password


@typechecked
def remove_all_except_git(dir_path: str):
    """
    Remove all in directory except .git/

    Args:
        dir_path (str): path to dir which will cleanup.
    """
    for file in os.listdir(dir_path):
        path = os.path.join(dir_path, file)
        try:
            if os.path.isfile(path) or \
                    os.path.islink(path):

                os.unlink(path)

            elif os.path.isdir(path):
                if path == os.path.join(dir_path, '.git'):
                    continue

                shutil.rmtree(path)

        except Exception as exception:  # pylint: disable=broad-except
            print(f'Failed to delete {path}. Reason: {exception}')


@typechecked
def choose_git_provider(repo: Repo) -> str:
    """
    Choose git provider.
    Try automatically, if git remotes specified. Otherwise - ask user.

    Args:
        repo (obj): git.Repo object
    Returns:
        git_provider string
    """
    if not repo.remotes:  # Remotes not exist in locally init repos
        git_provider = ask_user('choose_git_provider')
        return git_provider

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
def get_data_from_aws_config(login: str, password: str) -> {bool, object}:
    """
    Get and parse config from file

    Args:
        login (str): CLI argument provided by user.
        password (str): CLI argument provided by user.
    Returns:
        config ConfigParser obj|False: config, if exists. Otherwise - False
    """
    # Skip if CLI args provided
    if login and password:
        return False

    try:
        # Required mount: $HOME/.aws/:/home/cluster.dev/.aws/:ro

        # User can have multiply creds, ask him what should be used
        # https://boto3.amazonaws.com/v1/documentation/api/latest/guide/configuration.html#configuring-credentials
        file = f'{os.environ["HOME"]}/.aws/credentials'
        os.path.isfile(file)

        config = ConfigParser()
        config.read(file)

    except FileNotFoundError:
        return False

    return config


@typechecked
def get_aws_config_section(config: {object, bool}) -> str:
    """
    Ask user which section in config should be use for extracting credentials

    Args:
        config (ConfigParser obj|False): INI config
    Returns:
        config_section string: section name, if exists. Otherwise - empty string
    """
    # Skip if CLI args provided or configs doesn't exist
    if not config:
        return ''

    if len(config.sections()) == 0:
        return ''

    if len(config.sections()) == 1:
        section = config.sections()[0]
        print(f'Use AWS creds from file, section "{config.sections()[0]}"')
    else:
        section = ask_user('choose_config_section', choices=config.sections())

    return section


@typechecked
def get_aws_login(config: {object, bool}, config_section: str) -> str:
    """
    Get cloud programatic login from settings or from user input

    Args:
        config (ConfigParser obj|False): INI config
        config_section (str): INI config section
    Returns:
        cloud_login string
    """
    if not config:
        cloud_login = ask_user('cloud_login')
        return cloud_login

    cloud_login = config.get(config_section, 'aws_access_key_id')

    return cloud_login


@typechecked
def get_aws_password(config: {object, bool}, config_section: str) -> str:
    """
    Get cloud programatic password from settings or from user input

    Args:
        config (ConfigParser obj|False): INI config
        config_section (str): INI config section
    Returns:
        cloud_password string
    """

    if not config:
        cloud_password = ask_user('cloud_password')
        return cloud_password

    cloud_password = config.get(config_section, 'aws_secret_access_key')

    return cloud_password


@typechecked
def get_aws_session(config: {object, bool}, config_section: str, mfa_disabled: str = '') -> str:
    """
    Get cloud session from settings or from user input

    Args:
        config (ConfigParser obj|False): INI config
        config_section (str|False): INI config section
        mfa_disabled (str): CLI argument provided by user.
            If not provided - it set to empty string
    Returns:
        session_token string
    """
    # If login provided but session - not, user may not have MFA enabled
    if mfa_disabled:
        print('SESSION_TOKEN not found, try without MFA')
        return ''

    if not config:
        session_token = ask_user('aws_session_token')
        return session_token

    try:
        session_token = config.get(config_section, 'aws_session_token')
    except NoOptionError:
        print('SESSION_TOKEN not found, try without MFA')
        return ''

    return session_token


@typechecked
def create_aws_user_and_permissions(user: str, login: str, password: str, session: str) -> dict:
    """
    Create cloud user and attach needed permissions

    Args:
        user (str): Cluster.dev user name
        login (str): Cloud programatic login
        password (str): Cloud programatic password
        session (str): Cloud session token
    Returns:
        dict {'key': 'AWS_ACCESS_KEY_ID', 'secret': 'AWS_SECRET_ACCESS_KEY', 'created': bool}
    """
    iam = boto3.client(
        'iam',
        aws_access_key_id=login,
        aws_secret_access_key=password,
        aws_session_token=session,
    )
    keys_created = False
    # 'If specified user exist, -clogin and -cpwd will be try use as it
    # AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
    try:
        # Required "iam:ListAccessKeys"
        response = iam.list_access_keys(UserName=user)

        if response['AccessKeyMetadata'][0]['AccessKeyId'] == login:
            print(f'Specified creds belong to exist user {user}, use it')

            return {
                'key': login,
                'secret': password,
                'created': keys_created,
            }
    except iam.exceptions.NoSuchEntityException:
        pass

    try:  # Create/get policy_arn for user
        with open('aws_policy.json', 'r') as file:
            policy = json.dumps(json.load(file))

        response = iam.create_policy(
            PolicyName='cluster.dev-policy',
            Path='/cluster.dev/',
            PolicyDocument=policy,
            Description='Policy for https://cluster.dev propertly work. ' +
                        'Created by CLI instalator',
        )
        policy_arn = response['Policy']['Arn']
        print('Policy created')

    except iam.exceptions.EntityAlreadyExistsException:
        response = iam.list_policies(
            Scope='Local',
            PathPrefix='/cluster.dev/',
        )
        policy_arn = response['Policies'][0]['Arn']

    try:  # When cluster.dev user created, but provided creds for another user
        iam.get_user(UserName=user)

        response = iam.list_access_keys(UserName=user)
        key = response['AccessKeyMetadata'][0]['AccessKeyId']

        secret = ask_user('aws_secret_key', choices={'user': user, 'key': key})

    except iam.exceptions.NoSuchEntityException:
        # Create user and access keys
        iam.create_user(
            Path='/cluster.dev/',
            UserName=user,
        )
        iam.attach_user_policy(
            UserName=user,
            PolicyArn=policy_arn,
        )
        print('User created')

        response = iam.create_access_key(UserName=user)
        key = response['AccessKey']['AccessKeyId']
        secret = response['AccessKey']['SecretAccessKey']
        keys_created = True

    return {
        'key': key,
        'secret': secret,
        'created': keys_created,
    }


@typechecked
def get_git_token(provider: str) -> str:
    """
    Get github token from user input or settings

    Args:
        providercloud (str): Provider name
    Returns:
        git_token string
    """

    # Required env variables from host: GITHUB_TOKEN
    if provider == 'Github':
        if 'GITHUB_TOKEN' in os.environ:
            git_token = os.environ['GITHUB_TOKEN']
            print('Use GITHUB_TOKEN from env')
        else:
            git_token = ask_user('github_token')

    return git_token


@typechecked
def get_repo_name_from_url(url: str) -> str:
    """
    Get git repository name from git remote url
    """
    last_slash_index = url.rfind("/")
    last_suffix_index = url.rfind(".git")

    if last_slash_index < 0 or last_suffix_index < 0:
        _exit_(f'ERROR: Check `git remote -v`. Badly formatted origin: {url}')

    return url[last_slash_index + 1:last_suffix_index]


@typechecked
def get_repo_owner_from_url(url: str) -> str:
    """
    Get git repository owner from git remote url
    """
    last_slash_index = url.rfind("/")

    # IF HTTPS: https://github.com/shalb/cluster.dev.git
    prefix_index = url.find("/", 8)
    # IF SSH: git@github.com:shalb/cluster.dev.git
    if prefix_index == last_slash_index:
        prefix_index = url.find(":")

    if last_slash_index < 0 or prefix_index < 0:
        _exit_(f'ERROR: Check `git remote -v`. Badly formatted origin: {url}')

    return url[prefix_index + 1:last_slash_index]


@typechecked
def encrypt(public_key: str, secret_value: str) -> str:
    """Encrypt a Unicode string using the public key."""
    public_key = public.PublicKey(public_key.encode("utf-8"), encoding.Base64Encoder())
    sealed_box = public.SealedBox(public_key)
    encrypted = sealed_box.encrypt(secret_value.encode("utf-8"))
    return b64encode(encrypted).decode("utf-8")


@typechecked
def create_github_secrets(creds: dict, cloud: str, repo: Repo, git_token: str):
    """
    Create Github repo secrets for specified Cloud.

    Args:
        creds (dict(str)): Programmatic access to cloud
            { 'key': str, 'secret': str, ... }
        cloud (str): Cloud name
        repo (obj): git.Repo object
        git_token(str): Git Provider token
    """

    gh_api = GitHub(token=git_token)

    remote = repo.remotes.origin.url
    owner = get_repo_owner_from_url(remote)
    repo_name = get_repo_name_from_url(remote)

    # Get public key for encryption
    # https://developer.github.com/v3/actions/secrets/#get-a-repository-public-key
    for i in range(3, -1, -1):
        if i == 0:
            _exit_('ERROR: Can\'t access Github. Please, try again later')
        try:
            status, public_key = getattr(
                getattr(
                    getattr(
                        gh_api.repos, owner,
                    ), repo_name,
                ).actions.secrets, 'public-key',
            ).get()
            break
        except TimeoutError:
            print(f'Can\'t access Github. Timeout error. Attempts left: {i}')

    if status != 200:
        _exit_(f'Can\'t get repo public-key. Full error: {public_key}')

    if cloud == 'AWS':
        key = 'AWS_ACCESS_KEY_ID'
        secret = 'AWS_SECRET_ACCESS_KEY'

    # Put secret to repo
    # https://developer.github.com/v3/actions/secrets/#create-or-update-a-repository-secret
    body = {
        'encrypted_value': encrypt(public_key['key'], creds['key']),
        'key_id': public_key['key_id'],
    }

    for i in range(3, -1, -1):
        if i == 0:
            _exit_('ERROR: Can\'t access Github. Please, try again later')
        try:
            status, data = getattr(
                getattr(
                    getattr(
                        gh_api.repos, owner,
                    ), repo_name,
                ).actions.secrets, key,
            ).put(
                body=body,
            )
            break
        except TimeoutError:
            print(f'Can\'t access Github. Timeout error. Attempts left: {i}')

    if status not in (201, 204):
        _exit_(f'ERROR ocurred when try populate access_key to repo. {data}')

    body = {
        'encrypted_value': encrypt(public_key['key'], creds['secret']),
        'key_id': public_key['key_id'],
    }

    for i in range(3, -1, -1):
        if i == 0:
            _exit_('ERROR: Can\'t access Github. Please, try again later')
        try:
            status, data = getattr(
                getattr(
                    getattr(
                        gh_api.repos, owner,
                    ), repo_name,
                ).actions.secrets, secret,
            ).put(
                body=body,
            )
            break
        except TimeoutError:
            print(f'Can\'t access Github. Timeout error. Attempts left: {i}')

    if status not in (201, 204):
        _exit_(f'ERROR ocurred when try populate secret_key to repo. {data}')

    print('Secrets added to Github repo')


@typechecked
def get_last_release(owner: str = 'shalb', repo: str = 'cluster.dev') -> str:
    """
    Get last Github repo release

    Args:
        owner (str): repo owner name. Default: 'shalb'
        repo (str): repo name. Default: 'cluster.dev'
    Return:
        realise_version string
    """

    gh_api = GitHub()
    status, data = getattr(getattr(gh_api.repos, owner), repo).releases.latest.get()

    if status != 200:
        _exit_(f'Can\'t access Github. Full error: {data}')

    return data['tag_name']


@typechecked
def add_sample_cluster_dev_files(
        cloud: str,
        cloud_provider: str,
        git_provider: str,
        repo_path: str,
        release: str,
):
    """
    Populate repo with sample files

    Args:
        cloud (str): Cloud name
        cloud_provider (str): Cloud provider name
        git_provider (str): Git Provider name
        repo_path (str): Relative path to repo
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
    os.chmod(ci_file_path, 0o666)

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
    os.chmod(config_file_path, 0o666)

    print('\nLocal repo populated with sample files')


@typechecked
def commit_and_push(git: object, message: str):
    """
    Stage, commit and push all changes

    Args:
        git (obj): git.Repo.git object
        message(str): Commit message body
    """
    git.add('-A')
    try:
        git.commit('-m', message)
    except GitCommandError:
        print('Nothing to commit, working tree clean')

    git.push()

    print(message)


@typechecked
def last_edited_config_path(repo_path: str) -> str:
    """
    Get path to last edited .cluster.dev config

    Args:
        repo_path (str): Relative path to repo
    Return:
        last_changed_config_path string - Relative path to last edited config
    """
    # *.yaml means all files with .yaml extention
    config_mask = os.path.join(repo_path, '.cluster.dev', '*.yaml')
    config_files = glob.glob(config_mask)
    last_changed_config_path = max(config_files, key=os.path.getmtime)

    return last_changed_config_path


@typechecked
def set_cluster_installed(value: bool, config_path: str):
    """
    Set cluster.installed option in cluster config file

    Args:
        value (bool): Value that should be set to `installed` option
        config_path (str): Relative path to .cluster.dev config
    """

    # Change `installed` value
    prev_value = json.dumps(not value)
    new_value = json.dumps(value)

    with open(config_path, 'r') as conf:
        file = conf.read()

    new_file = file.replace(f'installed: {prev_value}', f'installed: {new_value}')

    with open(config_path, 'w') as conf:
        conf.write(new_file)


@typechecked
def main():
    """Logic"""

    cli = parse_cli_args()
    dir_path = os.path.relpath('current_dir')

    print('Hi, we gonna create an infrastructure for you.\n')

    if not dir_is_git_repo(dir_path):
        create_repo = True
        if not cli.repo_name:
            print('As this is a GitOps approach we need to start with the git repo.')
            create_repo = ask_user('create_repo')

        if not create_repo:
            _exit_('OK. See you soon!')

        repo_name = cli.repo_name or ask_user('repo_name')

        # TODO: setup remote origin and so on. Can be useful:
        # user = get_git_username(cli.git_user_name, git)
        # password = get_git_password(cli.git_password, git)
        _exit_('TODO')

    print('Inside git repo, use it.')

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
            pass

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

        creds = create_aws_user_and_permissions(cloud_user, access_key, secret_key, session_token)
        if creds['created']:
            print(
                f'Credentials for user "{cloud_user}":\n' +
                f'aws_access_key_id={creds["key"]}\n' +
                f'aws_secret_access_key={creds["secret"]}',
            )
    # elif cloud == 'DigitalOcean':
        # TODO
        # https://www.digitalocean.com/docs/apis-clis/doctl/how-to/install/
        # cloud_login = cli.cloud_login or get_do_login()
        # cloud_password = cli.cloud_password or get_do_password()
        # cloud_token = cli.cloud_token or ask_user('cloud_token')

        # create_cloud_user_and_req_permissions(cloud, cli.cloud_user, cloud_login, cloud_password)

    cloud_provider = cli.cloud_provider or \
        ask_user('choose_cloud_provider', choices=CLOUD_PROVIDERS[cloud])

    if git_provider == 'Github':
        create_github_secrets(creds, cloud, repo, git_token)

    release_version = cli.release_version or get_last_release()
    add_sample_cluster_dev_files(cloud, cloud_provider, git_provider, dir_path, release_version)
    commit_and_push(git, 'cluster.dev: Add sample files')

    config_path = last_edited_config_path(dir_path)
    set_cluster_installed(True, config_path)
    # Open editor
    os.system(f'editor "{config_path}"')
    commit_and_push(git, 'cluster.dev: Up cluster')

    # Show likn to logs. Temporary solution
    if git_provider == 'Github':
        remote = repo.remotes.origin.url
        owner = get_repo_owner_from_url(remote)
        repo_name = get_repo_name_from_url(remote)
        print(f'See logs at: https://github.com/{owner}/{repo_name}/actions')


#######################################################################
#                         G L O B A L   A R G S                       #
#######################################################################
GIT_PROVIDERS = ['Github']  # , 'Bitbucket', 'Gitlab']
CLOUDS = ['AWS', 'DigitalOcean']
CLOUD_PROVIDERS = {
    'AWS': [
        'Minikube',
        'AWS EKS',
    ],
    'DigitalOcean': [
        'Minikube',
        'Managed Kubernetes',
    ],
}

if __name__ == "__main__":
    # execute only if run as a script
    main()
