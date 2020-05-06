#!/usr/bin/env python3
"""
cluster.dev installation script

Example usage:
    ./cluster.dev.py install
    ./cluster.dev.py install -h
    ./cluster.dev.py install -p Github
"""
from __future__ import print_function
from __future__ import unicode_literals

import argparse
import os
import shutil
from sys import exit as _exit_

import regex
from git import GitCommandError
from git import InvalidGitRepositoryError
from git import Repo
from PyInquirer import prompt
from PyInquirer import style_from_dict
from PyInquirer import Token
from PyInquirer import ValidationError
from PyInquirer import Validator
# PyInquirer - Draw menu and user select one of items


#######################################################################
#                          F U N C T I O N S                          #
#######################################################################

def dir_is_git_repo(path):
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


class RepoNameValidator(Validator):  # pylint: disable=too-few-public-methods
    """Validate user input"""

    def validate(self, document=None, interactive=True):
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


class UserNameValidator(Validator):  # pylint: disable=too-few-public-methods
    """Validate user input"""

    def validate(self, document=None, interactive=True):
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


def ask_user(question_name, non_interactive_value=None, choices=None):
    """
    Draw menu for user interactions

    Args:
        question_name (str): Name from `questions` that will processed.
        non_interactive_value (str,int): used for skip interactive. Default to None
        choices (list): used when choose params become know during execution
    Returns:
        `non_interactive_value` if it set.
        Otherwise, depends on entered `question_name`:
            create_repo (bool): True or False
            git_provider (string): Provider name. See `GIT_PROVIDERS` for variants
            repo_name (string,int): Repo name. Have built-in validator,
                so user can't enter invalid repository name
    Raises:
        KeyError: If `question_name` not exist in `questions`
    """
    if non_interactive_value is not None:
        return non_interactive_value

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
    }

    try:
        result = prompt(questions[question_name], style=prompt_style)[question_name]
    except KeyError:
        _exit_(f'Sorry, it\'s program error. Can\'t found key "{question_name}"')

    return result


def parse_cli_args():
    """
    Parse CLI arguments, validate it
    """
    parser = argparse.ArgumentParser(
        usage='' +
        '  interactive:     ./cluster-dev.py install\n' +
        '  non-interactive: ./cluster-dev.py install -p Github',
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
        '--cloud', '-c', metavar='<cloud>',
        dest='cloud',
        help="Can be 'Amazon Web Services' or 'Digital Ocean'",
        choices=CLOUDS,
    )
    parser.add_argument(
        '--cloud-provider', '-cp', metavar='<provider>',
        dest='cloud_provider',
        help='Cloud provider depends on selected --cloud',
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


def get_git_username(cli_arg, git):
    """
    Get username from settings or from user input

    Args:
        cli_arg (str): CLI argument provided by user.
            If not provided - it set to None
        git (obj): git.Repo.git object
    Returns:
        username string
    """
    if cli_arg:
        return cli_arg

    try:
        # Required mount: $HOME/.gitconfig:/home/cluster.dev/.gitconfig:ro
        user = git.config('--get', 'user.name')
        print(f'Username: {user}')
    except GitCommandError:
        user = ask_user('git_user_name')

    return user


def get_git_password(cli_arg):
    """
    Get SSH key from settings or password from user input

    Args:
        cli_arg (str): CLI argument provided by user.
            If not provided - it set to None
    Returns:
        `False` if SSH-key provided (mounted)
        Otherwise - return password string
    """
    # If password provided - return password
    if cli_arg:
        return cli_arg

    try:
        # Try use ssh as password
        # Required mount: $HOME/.ssh:/home/cluster.dev/.ssh:ro
        if os.listdir(f'{os.environ["HOME"]}/.ssh'):
            print('Password type: ssh-key')
            return False

    except FileNotFoundError:
        pass

    password = ask_user('git_password')
    return password


def remove_all_except_git(dir_path):
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


def choose_git_provider(cli_arg, repo):
    """
    Choose git provider.
    Try automatically, if git remotes specified. Otherwise - ask user.

    Args:
        cli_arg (str): CLI argument provided by user.
            If not provided - it set to None
        repo (obj): git.Repo object
    Returns:
        git_provider string
    """
    # If git provider provided - return git provider
    if cli_arg:
        return cli_arg

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


def choose_cloud(cli_arg):
    """
    Get cloud from cli or from user input

    Args:
        cli_arg (str): CLI argument provided by user.
            If not provided - it set to None
    Returns:
        cloud string
    """
    # If cloud provided - return cloud string
    if cli_arg:
        return cli_arg

    cloud = ask_user('choose_cloud')

    return cloud


def choose_cloud_provider(cli_arg, cloud_providers):
    """
    Choose git provider.
    Try automatically, if git remotes specified. Otherwise - ask user.

    Args:
        cli_arg (str): CLI argument provided by user.
            If not provided - it set to None
        cloud_providers (list)
    Returns:
        cloud_provider string
    """
    # If cloud provided - return cloud string
    if cli_arg:
        return cli_arg

    cloud_provider = ask_user('choose_cloud_provider', choices=cloud_providers)

    return cloud_provider


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

        repo_name = ask_user('repo_name', cli.repo_name)

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

    git_provider = choose_git_provider(cli.git_provider, repo)

    if not repo.remotes:
        publish_repo = ask_user('publish_repo_to_git_provider')
        if publish_repo:
            # TODO: push repo to Git Provider
            pass

    user = get_git_username(cli.git_user_name, git)
    password = get_git_password(cli.git_password)

    cloud = choose_cloud(cli.cloud)
    cloud_provider = choose_cloud_provider(cli.cloud_provider, CLOUD_PROVIDERS[cloud])


#######################################################################
#                         G L O B A L   A R G S                       #
#######################################################################
GIT_PROVIDERS = ['Github', 'Bitbucket', 'Gitlab']
CLOUDS = ['Amazon Web Services', 'Digital Ocean']
CLOUD_PROVIDERS = {
    'Amazon Web Services': [
        'Minikube',
        'AWS EKS',
    ],
    'Digital Ocean': [
        'Minikube',
        'Managed Kubernetes',
    ],
}

if __name__ == "__main__":
    # execute only if run as a script
    main()
