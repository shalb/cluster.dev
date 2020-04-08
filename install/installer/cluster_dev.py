#!/usr/bin/env python3

"""
cluster.dev installation script

Example usage:
    ./cluster.dev.py install
    ./cluster.dev.py install -h
    ./cluster.dev.py install -p Github
"""

from __future__ import print_function, unicode_literals
import argparse
import os
from git import Repo, InvalidGitRepositoryError, GitCommandError
import regex
# PyInquirer - Draw menu and user select one of items
from PyInquirer import style_from_dict, Token, prompt, Separator, Validator, ValidationError


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


class RepoNameValidator(Validator):
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
                exit(error_message)

            raise ValidationError(
                message=error_message,
                cursor_position=len(repo_name))  # Move cursor to end


class UserNameValidator(Validator):
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
                exit(error_message)

            raise ValidationError(
                message=error_message,
                cursor_position=len(username))  # Move cursor to end


def ask_user(question_name, non_interactive_value=None):
    """
    Draw menu for user interactions

    Args:
        question_name (str): Name from `questions` that will processed.
        non_interactive_value (str,int): used for skip interactive. Default to None
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
        #Token.Selected: '',  # default
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
            'choices': [{
                'name': 'No, I create or clone repo and then run tool there',
                'value': False
            }, {
                'name': 'Yes',
                'value': True,
                'disabled': 'Unavailable at this time',
            }, ]
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
    }

    try:
        result = prompt(questions[question_name], style=prompt_style)[question_name]
    except KeyError:
        exit(f'Sorry, it\'s program error. Can\'t found key "{question_name}"')

    return result


def parse_cli_args():
    """
    Parse CLI arguments, validate it
    """
    parser = argparse.ArgumentParser(
        usage='' + \
        '  interactive:     ./cluster-dev.py install\n' +
        '  non-interactive: ./cluster-dev.py install -p Github')

    parser.add_argument('subcommand', nargs='+', metavar='install',
                        choices=['install'])
    parser.add_argument('--git-provider', '-p', metavar='<provider>',
                        dest='git_provider',
                        help='Can be Github, Bitbucket or Gitlab',
                        choices=GIT_PROVIDERS)
    parser.add_argument('--create-repo', metavar='<repo_name>',
                        dest='repo_name',
                        help='Automatically initialize repo for you')
    parser.add_argument('--git-user-name', metavar='<username>',
                        dest='git_user_name',
                        help='Username used in Git Provider.' + \
                            'Can be automatically get from .gitconfig')
    parser.add_argument('--git-password', metavar='<password>',
                        dest='git_password',
                        help='Password used in Git Provider. ' + \
                            'Can be automatically get from .ssh')
    cli = parser.parse_args()

    if cli.repo_name:
        RepoNameValidator.validate(cli.repo_name, interactive=False)

    if cli.git_user_name:
        UserNameValidator.validate(cli.git_user_name, interactive=False)

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
    user = cli_arg
    if not user:
        try:
            # Required mount: $HOME/.gitconfig:/home/cluster.dev/.gitconfig:ro
            user = git.config('--get', 'user.name')
            print(f'Username: {user}')
        except GitCommandError:
            user = ask_user('git_user_name')

    return user


def get_git_password(cli_arg, git):
    """
    Get SSH key from settings or password from user input

    Args:
        cli_arg (str): CLI argument provided by user.
            If not provided - it set to None
        git (obj): git.Repo.git object
    Returns:
        `False` if SSH-key provided (mounted)
        Otherwise - return password string
    """
    password = cli_arg

    if not password:
        try:
            use_ssh = git.remote('-v').find('git@')

            if use_ssh != -1:
                # Try use ssh as password
                # Required mount: $HOME/.ssh:/home/cluster.dev/.ssh:ro
                print('Password type: ssh-key')
                return False
        except GitCommandError:
            pass

        password = ask_user('git_password')

    return password


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
            exit('OK. See you soon!')

        repo_name = ask_user('repo_name', cli.repo_name)

        # TODO: setup remote origin and so on. Can be useful:
        # user = get_git_username(cli.git_user_name, git)
        # password = get_git_password(cli.git_password, git)
        exit('TODO')


    print('Inside git repo, use it.')

    repo = Repo(dir_path)
    git = repo.git


    if repo.heads: # Heads exist only after first commit
        cleanup_repo = ask_user('cleanup_repo')
        if cleanup_repo:
            # TODO: rm -rf (except .git) and commit
            pass


    # Choose git provider
    git_provider = cli.git_provider
    if not git_provider:
        if not repo.remotes: # Remotes not exist in locally init repos
            git_provider = ask_user('choose_git_provider')
            publish_repo = ask_user('publish_repo_to_git_provider')
            if publish_repo:
                # TODO: push repo to Git Provider
                pass
        else:
            remote = git.remote('-v')

            if remote.find('github'):
                git_provider = 'Github'
            elif remote.find('bitbucket'):
                git_provider = 'Bitbucket'
            elif remote.find('gitlab'):
                git_provider = 'Gitlab'
            else:
                git_provider = ask_user('choose_git_provider')


    user = get_git_username(cli.git_user_name, git)
    password = get_git_password(cli.git_password, git)


#######################################################################
#                         G L O B A L   A R G S                       #
#######################################################################

GIT_PROVIDERS = ['Github', 'Bitbucket', 'Gitlab']

if __name__ == "__main__":
    # execute only if run as a script
    main()
