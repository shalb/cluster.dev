#!/usr/bin/env python3
"""AWS interaction functions."""
import json
import logging
import sys
from typing import Dict
from typing import Union

import boto3
from validate import ask_user


try:
    from typeguard import typechecked  # noqa: WPS433
except ModuleNotFoundError:
    def typechecked(func=None):  # noqa: WPS440
        """Skip runtime type checking on the function arguments."""
        return func

logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)
logger.addHandler(logging.StreamHandler(sys.stdout))


@typechecked
def user_exist(iam: boto3.client, user: str, rights_msg: str) -> bool:
    """Check is specified AWS user exist.

    Required "iam:GetUser"

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
        sys.exit(f'\nERROR: {error}{rights_msg}')

    logger.debug(f'User "{user}" exist')
    return True


@typechecked
def list_user_access_keys(
    iam: boto3.client,
    user: str,
    rights_msg: str,
) -> list:
    """Get exist access keys for exist user.

    Required "iam:ListUsers"???, "iam:ListAccessKeys"

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
        sys.exit(f'\nERROR: {error}{rights_msg}')

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
) -> Dict[str, Union[str, bool]]:
    """Check that provided keys affiliate to provided username of exist user.

    If specified user exist, --cloud-programatic-login and
    --cloud-programatic-password will be try use as it.

    Args:
        user: (str) Cluster.dev user name.
        login: (str) Cloud programatic login.
        password: (str) Cloud programatic password.
        keys: (list) AWS_ACCESS_KEY_ID's
    Returns:
        Dict[str, Union[str, bool]]:
        - `{'key': 'AWS_ACCESS_KEY_ID', 'secret': 'AWS_SECRET_ACCESS_KEY', 'created': bool}`
        - `{}` - if keys not affiliate to user.
    """
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
def create_access_key(
    iam: boto3.client,
    user: str,
) -> Dict[str, Union[str, bool]]:
    """Create Access Key for exist user.

    Required "iam:CreateAccessKey".

    Args:
        iam: (boto3.client.IAM) A low-level client representing AWS IAM.
        user: (str) Cluster.dev user name.

    Returns:
        Dict[str, Union[str, bool]]:
        `{'key': 'AWS_ACCESS_KEY_ID', 'secret': 'AWS_SECRET_ACCESS_KEY', 'created': bool}`
    """
    response = iam.create_access_key(UserName=user)

    logger.debug('Access keys created')
    return {
        'key': response['AccessKey']['AccessKeyId'],
        'secret': response['AccessKey']['SecretAccessKey'],
        'created': True,
    }


@typechecked
def create_user_and_permissions(
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

    if user_exist(iam, user, msg):
        keys = list_user_access_keys(iam, user, msg)

        if not keys:
            return create_access_key(iam, user)

        creds = check_affiliation_user_and_creds(user, login, password, keys)
        if creds:
            return creds

    keys_created = False
    link = (
        f'https://github.com/shalb/cluster.dev/blob/{release}'
        + '/install/installer_aws_install_req_permissions.json'
    )

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

        secret = ask_user(
            name='aws_secret_key',
            type='password',
            message=f"Please enter AWS Secret Key for:\n  User: '{user}'\n  Public Key: '{key}'\n",
        )

    return {
        'key': key,
        'secret': secret,
        'created': keys_created,
    }
