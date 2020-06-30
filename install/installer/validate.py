#!/usr/bin/env python3
"""User input validators."""
import sys

import regex
from PyInquirer import prompt
from PyInquirer import style_from_dict
from PyInquirer import Token
from PyInquirer import ValidationError
from PyInquirer import Validator


try:
    from typeguard import typechecked  # noqa: WPS433
except ModuleNotFoundError:
    def typechecked(func=None):  # noqa: WPS440
        """Skip runtime type checking on the function arguments."""
        return func


@typechecked
class RepoName(Validator):  # pylint: disable=too-few-public-methods
    """Validator for the Github repository name input."""

    @typechecked
    def validate(self: Validator, document=None, interactive: bool = True):
        """Verify that the user input string conforms Github username requirements.

        Args:
            document: (prompt_toolkit.document.Document)
                Contain validation data in interactive mode.
                When not interactive, not used and set to None.
            interactive: (bool) Work mode.
                Set to False when check CLI params input, Default to True.

        Raises:
            ValidationError: If intput string not match regex.
                Show error message for user and get him change fix error.
        """
        repo_name = self
        if interactive:
            repo_name = document.text

        error_message = (
            'Please enter a valid repository name. It should have 1-100 '
            + 'only latin letters, numbers, underscores and dashes'
        )

        okay = regex.match('^[A-Za-z0-9_-]{1,100}$', repo_name)

        if okay:
            return

        if not interactive:
            sys.exit(error_message)

        raise ValidationError(
            message=error_message,
            cursor_position=len(repo_name),
        )  # Move cursor to end


@typechecked
class UserName(Validator):  # pylint: disable=too-few-public-methods
    """Validator for the Github username input."""

    @typechecked
    def validate(self: Validator, document=None, interactive: bool = True):
        """Verify that the user input string conforms Github repo name requirements.

        Args:
            document: (prompt_toolkit.document.Document)
                Contain validation data in interactive mode.
                When not interactive, not used and set to None.
            interactive: (bool) Work mode.
                Set to False when check CLI params input, Default to True.

        Raises:
            ValidationError: If intput string not match regex.
                Show error message for user and get him change fix error.
        """
        username = self
        if interactive:
            username = document.text

        error_message = (
            'Please enter a valid username. '
            + 'It should have 1-39 only latin letters, numbers, '
            + 'single hyphens and cannot begin or end with a hyphen'
        )

        okay = regex.match('^(?!-)[A-Za-z0-9-]{1,39}$', username)
        not_ok = regex.match('^.*[-]{2,}.*$', username)
        not_ok2 = regex.match('^.*-$', username)

        if not_ok or not_ok2 or not okay:
            if not interactive:
                sys.exit(error_message)

            raise ValidationError(
                message=error_message,
                cursor_position=len(username),
            )  # Move cursor to end


@typechecked
class AWSUserName(Validator):  # pylint: disable=too-few-public-methods
    """Validator for the AWS username input."""

    @typechecked
    def validate(self: Validator, document=None, interactive: bool = True):
        """Verify that the user input string conforms AWS username requirements.

        Args:
            document: (prompt_toolkit.document.Document)
                Contain validation data in interactive mode.
                When not interactive, not used and set to None.
            interactive: (bool) Work mode.
                Set to False when check CLI params input, Default to True.

        Raises:
            ValidationError: If intput string not match regex.
                Show error message for user and get him change fix error.
        """
        username = self
        if interactive:
            username = document.text

        error_message = (
            'Please enter a valid AWS username. It should have 1-64 '
            + 'only latin letters, numbers and symbols: _+=,.@-'
        )

        okay = regex.match('^[A-Za-z0-9_+=,.@-]{1,64}$', username)

        if okay:
            return

        if not interactive:
            sys.exit(error_message)

        raise ValidationError(
            message=error_message,
            cursor_position=len(username),
        )  # Move cursor to end


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
            'choices': choices,
        },
        'repo_name': {
            'type': 'input',
            'name': 'repo_name',
            'message': 'Please enter the name of your infrastructure repository',
            'default': 'infrastructure',
            'validate': RepoName,
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
            'validate': UserName,
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
            'choices': choices,
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
            'validate': AWSUserName,
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
