#!/usr/bin/env python3
"""AWS interaction functions."""
import json
import sys
from typing import Dict
from typing import Literal
from typing import Union

import boto3
from logger import logger
from validate import ask_user


try:
    from typeguard import typechecked  # noqa: WPS433
except ModuleNotFoundError:
    def typechecked(func=None):  # noqa: WPS440
        """Skip runtime type checking on the function arguments."""
        return func


FalseL = Literal[False]
TrueL = Literal[True]


@typechecked
def user_exist(iam: boto3.client, user: str, rights_msg: str) -> bool:
    """Check is specified AWS user exist.

    Required "iam:GetUser".

    Args:
        iam: (boto3.client.IAM) A low-level client representing AWS IAM.
        user: (str) Cluster.dev user name.
        rights_msg: (str) Notification message of lack of rights.

    Returns:
        bool - True if user exist and False if not.
    """
    try:
        iam.get_user(UserName=user)
    except iam.exceptions.NoSuchEntityException:
        logger.debug(f'User "{user}" does not exist yet')
        return False

    except iam.exceptions.ClientError as error:
        logger.error(f'{error}{rights_msg}')
        sys.exit()

    logger.debug(f'User "{user}" exist')
    return True


@typechecked
def get_user_access_keys(
    iam: boto3.client,
    user: str,
    rights_msg: str,
) -> list:
    """Get exist access keys for exist user.

    Required "iam:ListUsers"???, "iam:ListAccessKeys".

    Args:
        iam: (boto3.client.IAM) A low-level client representing AWS IAM.
        user: (str) Cluster.dev user name.
        rights_msg: (str) Notification message of lack of rights.

    Returns:
        list - AWS_ACCESS_KEY_ID's.
    """
    try:
        response = iam.list_access_keys(UserName=user)

    except iam.exceptions.ClientError as error:
        logger.error(f'{error}{rights_msg}')
        sys.exit()

    keys = []
    for key in response['AccessKeyMetadata']:
        keys.append(key['AccessKeyId'])

    logger.debug(f'User access keys: {keys}')
    return keys


@typechecked
def check_affiliation_user_and_creds(
    user: str,
    login: str,
    password: str,
    keys: list,
) -> Dict[str, Union[str, FalseL]]:
    """Check that provided keys affiliate to provided username of exist user.

    If specified user exist, `--cloud-programatic-login` and
    `--cloud-programatic-password` will be try use as it.

    Args:
        user: (str) Cluster.dev user name.
        login: (str) Cloud programatic login.
        password: (str) Cloud programatic password.
        keys: (list) AWS_ACCESS_KEY_ID's.

    Returns:
        Dict[str, Union[str, bool]]:
        * `{'key': 'AWS_ACCESS_KEY_ID', 'secret': 'AWS_SECRET_ACCESS_KEY', 'created': False}`
        Or `{}` - if keys not affiliate to user.
    """  # noqa: P102
    for key in keys:
        if key == login:
            logger.info(f'Specified creds belong to user "{user}", use it')

            return {
                'key': login,
                'secret': password,
                'created': False,
            }

    logger.debug(f'Specified keys not belong to user "{user}"')
    return {}


@typechecked
def create_access_key(  # noqa: WPS234
    iam: boto3.client,
    user: str,
    rights_msg: str,
) -> Dict[str, Union[str, TrueL]]:
    """Create Access Key for exist user.

    Required "iam:CreateAccessKey".

    Args:
        iam: (boto3.client.IAM) A low-level client representing AWS IAM.
        user: (str) Cluster.dev user name.
        rights_msg: (str) Notification message of lack of rights.

    Returns:
        Dict[str, Union[str, bool]]:
        `{'key': 'AWS_ACCESS_KEY_ID', 'secret': 'AWS_SECRET_ACCESS_KEY', 'created': True}`
    """
    try:
        response = iam.create_access_key(UserName=user)
    except iam.exceptions.LimitExceededException:
        logger.error(
            f"User '{user}' already have 2 programmatic keys.\n"
            + 'If you have access to one of it - provide it credentials. Or remove unused key.\n'
            + f'https://console.aws.amazon.com/iam/home#/users/{user}?section=security_credentials',
        )
        sys.exit()

    except iam.exceptions.ClientError as error:
        logger.error(f'{error}{rights_msg}')
        sys.exit()

    logger.debug('Access keys created')
    return {
        'key': response['AccessKey']['AccessKeyId'],
        'secret': response['AccessKey']['SecretAccessKey'],
        'created': True,
    }


@typechecked
def create_user(
    iam: boto3.client,
    username: str,
    rights_msg: str,
    path: str = '/cluster.dev/',
):
    """Create AWS IAM user.

    Required "iam:CreateUser".

    Args:
        iam: (boto3.client.IAM) A low-level client representing AWS IAM.
        username: (str) Cluster.dev user name.
        rights_msg: (str) Notification message of lack of rights.
        path: (str) The path for the policy. Default: '/cluster.dev/'.
    """
    try:
        iam.create_user(
            Path=path,
            UserName=username,
        )
    except iam.exceptions.ClientError as error:
        logger.error(f'{error}{rights_msg}')
        sys.exit()

    logger.debug(f"User '{path}{username}' created")


@typechecked
def get_policy_arn(
    iam: boto3.client,
    rights_msg: str,
    prefix: str = '/cluster.dev/',
    policy_name: str = 'cluster.dev-policy',
) -> str:
    """Get policy ARN if exist.

    Required "iam:ListPolicies".

    Args:
        iam: (boto3.client.IAM) A low-level client representing AWS IAM.
        rights_msg: (str) Notification message of lack of rights.
        prefix: (str) The path prefix for filtering the results. Default: '/cluster.dev/'.
        policy_name: (str) Name of policy for which it will be necessary
            to find ARN. Default: 'cluster.dev-policy'.

    Returns:
        string - Policy ARN. If policy not exist - return empty string.
    """
    try:
        response = iam.list_policies(
            Scope='Local',
            PathPrefix=prefix,
        )
    except iam.exceptions.ClientError as error:
        logger.error(f'{error}{rights_msg}')
        sys.exit()

    for policy in response['Policies']:
        if policy_name == policy['PolicyName']:
            logger.debug(
                f"Policy ARN for '{prefix}{policy_name}' is '{policy['Arn']}'",
            )  # noqa: WPS221
            return policy['Arn']

    logger.debug(f"Policy '{prefix}{policy_name}' not exist yet.")
    return ''


@typechecked
def create_policy(
    iam: boto3.client,
    rights_msg: str,
    path: str = '/cluster.dev/',
    policy_name: str = 'cluster.dev-policy',
) -> str:
    """Create cluster.dev policy.

    Required "iam:CreatePolicy".

    Args:
        iam: (boto3.client.IAM) A low-level client representing AWS IAM.
        rights_msg: (str) Notification message of lack of rights.
        path: (str) The path for the policy. Default: '/cluster.dev/'.
        policy_name: (str) Policy name for creation. Default: 'cluster.dev-policy'.

    Returns:
        string - Created policy ARN.
    """
    # TODO Refactor: https://github.com/shalb/cluster.dev/pull/70#discussion_r449274402
    with open('aws_policy.json', 'r') as policy_file:
        policy = json.dumps(json.load(policy_file))

    try:
        response = iam.create_policy(
            PolicyName=policy_name,
            Path=path,
            PolicyDocument=policy,
            Description='Policy for https://cluster.dev propertly work. '
            + 'Created by CLI instalator',
        )
    except iam.exceptions.ClientError as error:
        logger.error(f'{error}{rights_msg}')
        sys.exit()

    logger.info('Policy created')
    logger.debug(
        f"Policy ARN for '{path}{policy_name}' is '{response['Policy']['Arn']}'",
    )  # noqa: WPS221

    return response['Policy']['Arn']


@typechecked
def attach_policy_to_user(iam: boto3.client, username: str, policy_arn: str, rights_msg: str):
    """Attach IAM policy to user.

    Required "iam:AttachUserPolicy".

    Args:
        iam: (boto3.client.IAM) A low-level client representing AWS IAM.
        username: (str) Cluster.dev user name.
        policy_arn: (str) Policy ARN.
        rights_msg: (str) Notification message of lack of rights.
    """
    try:
        iam.attach_user_policy(
            UserName=username,
            PolicyArn=policy_arn,
        )
    except iam.exceptions.ClientError as error:
        logger.error(f'{error}{rights_msg}')
        sys.exit()

    logger.debug(f"Attached policy '{policy_arn}' to user '{username}'")


@typechecked
def ask_user_for_provide_keys(
    user: str,
    login: str,
    keys: list,
) -> Dict[str, Union[str, FalseL]]:
    """Give user chance to specify right keys without program restart.

    Args:
        user: (str) Cluster.dev user name.
        login: (str) Cloud programatic login.
        keys: (list) WS_ACCESS_KEY_ID's.

    Returns:
        Dict[str, Union[str, bool]]:
        `{'key': 'AWS_ACCESS_KEY_ID', 'secret': 'AWS_SECRET_ACCESS_KEY', 'created': False}`
    """
    have_secret = ask_user(
        type='confirm',
        message=f'User {user} have 2 programmatic keys.\n'
        + f"Key '{login}' used for IAM interactions not belong to user '{user}'\n"
        + f'that have: {keys} and it maximum supply.\n\n'
        + 'Would you have Secret Key for on of this keys?',
        default=False,
    )

    if not have_secret:
        sys.exit(
            'Well, you need remove unused key and then try again.\n'
            + f'https://console.aws.amazon.com/iam/home#/users/{user}'
            + '?section=security_credentials',
        )

    key = ask_user(
        type='list',
        message='Select key for what you known AWS_SECRET_ACCESS_KEY',
        choices=keys,
    )
    secret = ask_user(
        type='password',
        message='Please enter AWS_SECRET_ACCESS_KEY:',
        # TODO: add validator
    )

    return {
        'key': key,
        'secret': secret,
        'created': False,
    }


@typechecked
def create_user_and_permissions(  # noqa: WPS211
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

    # cli_rights_needed_msg
    msg = (
        '\n\nRights what you need described here: \n'
        + f'https://github.com/shalb/cluster.dev/blob/{release}'
        + '/install/installer_aws_install_req_permissions.json'
    )

    #
    # Get user AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
    #
    if user_exist(iam, user, msg):
        keys = get_user_access_keys(iam, user, msg)

        if not keys:  # noqa: WPS504
            creds = create_access_key(iam, user, msg)
        else:
            # Check is provided keys belong to user.
            # Return creds if affiliation confirmed.
            creds = check_affiliation_user_and_creds(user, login, password, keys)
            if not creds:
                if len(keys) < 2:
                    creds = create_access_key(iam, user, msg)
                else:
                    # Give user chance to specify right keys without program restart.
                    creds = ask_user_for_provide_keys(user, login, keys)

    else:
        create_user(iam, user, msg)
        creds = create_access_key(iam, user, msg)

    #
    # Attach policy to user
    #
    policy_arn = get_policy_arn(iam, msg)
    if not policy_arn:
        policy_arn = create_policy(iam, msg)

    attach_policy_to_user(iam, user, policy_arn, msg)

    return creds
