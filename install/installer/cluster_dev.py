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
import regex
# PyInquirer - Draw menu and user select one of items
from PyInquirer import style_from_dict, Token, prompt, Separator, Validator, ValidationError

from git import Repo, InvalidGitRepositoryError

#######################################################################
#                          F U N C T I O N S                          #
#######################################################################

def current_dir_is_git_repo():
    """
    Check if current directory is a Git repository

    Returns:
        bool: True if successful, False otherwise.
    """
    try:
        Repo(os.path.relpath('current_dir'))
        return True
    except InvalidGitRepositoryError:
        return False

class RepoNameValidator(Validator):
    """Validate user input"""
    def validate(self, document):
        """
        Validate user input string to Github repo name restrictions

        Raises:
            ValidationError: If intput string not match regex.
                Show error message for user and get him change fix error
        """
        match = regex.match('^.{1,100}$', document.text)
        if not match:
            raise ValidationError(
                message='Please enter a valid repository name. ' + \
                'It should have from 1 to 100 latin letters, ' + \
                'numbers, underscores or dashes',
                cursor_position=len(document.text))  # Move cursor to end

        match = regex.match('^[A-Za-z0-9_-]+$', document.text)
        if not match:
            raise ValidationError(
                message='Please enter a valid repository name. ' + \
                'It can include only latin letters, numbers, ' + \
                'underscores and dashes.',
                cursor_position=len(document.text))  # Move cursor to end



def interactive_input(question_name):
    """
    Draw menu for user interactions

    Args:
        question_name (str): Name from `questions` that will processed.

    Returns:
        Depends on entered parameter:
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
        'git_provider': {
            'type': 'list',
            'name': 'git_provider',
            'message': 'Select your Git hosting',
            'choices': GIT_PROVIDERS,
        },
        'repo_name': {
            'type': 'input',
            'name': 'repo_name',
            'message': 'Please enter the name of your infrastructure repository',
            'default': 'infrastructure',
            'validate': RepoNameValidator,
        },
    }

    try:
        result = prompt(questions[question_name], style=prompt_style)[question_name]
    except KeyError:
        exit(f'Sorry, it\'s program error. Can\'t found key "{question_name}"')

    return result


#######################################################################
#                         G L O B A L   A R G S                       #
#######################################################################

GIT_PROVIDERS = ['Github', 'Bitbucket', 'Gitlab']


#######################################################################
#                      P A R S E   C L I   A R G S                    #
#######################################################################

PARSER = argparse.ArgumentParser(usage='''
  interactive:     ./cluster-dev.py install
  non-interactive: ./cluster-dev.py install -p Github
''')

PARSER.add_argument('subcommand', nargs='+', metavar='install',
                    choices=['install'])
PARSER.add_argument("--git_provider", "-p", metavar='<provider>',
                    help='Can be Github, Bitbucket or Gitlab',
                    choices=GIT_PROVIDERS)
ARGS = PARSER.parse_args()


#######################################################################
#                               L O G I C                             #
#######################################################################

print('Hi, we gonna create an infrastructure for you.\n')


if not current_dir_is_git_repo():
    print('As this is a GitOps approach we need to start with the git repo.')

    CREATE_REPO = interactive_input('create_repo')

    if not CREATE_REPO:
        exit("OK. See you soon!")


    REPO_NAME = interactive_input('repo_name')

    # TODO: get credentials from file
    # TODO: https://gitpython.readthedocs.io/en/stable/intro.html



print('Inside git repo, use it.')

GIT_PROVIDER = ARGS.git_provider
if not GIT_PROVIDER:
    GIT_PROVIDER = interactive_input('git_provider')
