#!/usr/bin/env python3
"""User input validators."""
import sys
from typing import Any
from typing import Union

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
def ask_user(**kwargs: Any) -> Union[str, bool]:
    """Draw menu for interactions with user.

    See https://github.com/CITGuru/PyInquirer#examples for argument names and it types.
    If you not specify 'name' field - it will be get from 'message'.
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

    name = kwargs.get('name') or kwargs.get('message')
    question = {
        'name': name,
    }

    for key, value in kwargs.items():  # noqa: WPS110
        question[key] = value

    try:
        answer = prompt(question, style=prompt_style)[name]
    except KeyError:
        sys.exit(f"Sorry, it's program error. Can't execute '{question}'")

    return answer
