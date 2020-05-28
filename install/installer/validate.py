#!/usr/bin/env python3
"""User input validators."""
import sys

import regex
from PyInquirer import ValidationError
from PyInquirer import Validator
from typeguard import typechecked


@typechecked  # pylint: disable=too-few-public-methods
class RepoName(Validator):
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


@typechecked  # pylint: disable=too-few-public-methods
class UserName(Validator):
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


@typechecked  # pylint: disable=too-few-public-methods
class AWSUserName(Validator):
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
