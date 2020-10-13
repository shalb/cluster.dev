#!/usr/bin/env python3
"""Interaction functions with local user AWS config."""
import os
from configparser import ConfigParser
from configparser import NoOptionError
from typing import Literal
from typing import Union

from logger import logger
from validate import ask_user


try:
    from typeguard import typechecked  # noqa: WPS433
except ModuleNotFoundError:
    def typechecked(func=None):  # noqa: WPS440
        """Skip runtime type checking on the function arguments."""
        return func


@typechecked
def parse_file(login: str, password: str) -> Union[ConfigParser, Literal[False]]:  # noqa: WPS212
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
def choose_section(config: Union[ConfigParser, Literal[False]]) -> str:
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
        section = ask_user(
            name='choose_config_section',
            type='list',
            message='Select credentials section that will be used to deploy cluster.dev',
            choices=config.sections(),
        )

    return section


@typechecked
def get_login(config: Union[ConfigParser, Literal[False]], config_section: str) -> str:
    """Get cloud programatic login from settings or from user input.

    Args:
        config (configparser.ConfigParser|False): INI config.
        config_section: (str) INI config section.

    Returns:
        cloud_login string.
    """
    if not config or not config_section:
        # TODO: Add validation (and for cli arg)
        return ask_user(
            name='cloud_login',
            type='input',
            message='Please enter your Cloud programatic key',
        )

    return config.get(config_section, 'aws_access_key_id')


@typechecked
def get_password(config: Union[ConfigParser, Literal[False]], config_section: str) -> str:
    """Get cloud programatic password from settings or from user input.

    Args:
        config: (configparser.ConfigParser|False) INI config.
        config_section: (str) INI config section.

    Returns:
        cloud_password string.
    """
    if not config:
        # TODO: Add validation (and for cli arg)
        return ask_user(
            name='cloud_password',
            type='password',
            message='Please enter your Cloud programatic secret',
        )

    return config.get(config_section, 'aws_secret_access_key')


@typechecked
def get_session(  # noqa: WPS212
    config: Union[ConfigParser, Literal[False]],
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
        return ask_user(
            name='aws_session_token',
            type='password',
            message='Please enter your AWS Session token',
        )

    try:
        session_token = config.get(config_section, 'aws_session_token')
    except NoOptionError:
        logger.info('SESSION_TOKEN not found, try without MFA')
        return ''

    return session_token
