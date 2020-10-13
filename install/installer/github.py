#!/usr/bin/env python3
"""Github interaction functions."""
import sys
from base64 import b64encode
from typing import Dict
from typing import Union

from agithub.GitHub import GitHub
from logger import logger
from nacl import encoding
from nacl import public

try:
    from typeguard import typechecked  # noqa: WPS433
except ModuleNotFoundError:
    def typechecked(func=None):  # noqa: WPS440
        """Skip runtime type checking on the function arguments."""
        return func


@typechecked
def encrypt(public_key: str, secret_value: str) -> str:
    """Encrypt a Unicode string using the public key.

    Args:
        public_key: (str) encryption salt public key.
        secret_value: (str) value for encrypt.

    Returns:
        Encrypted Unicode string.
    """
    public_key = public.PublicKey(public_key.encode('utf-8'), encoding.Base64Encoder())
    sealed_box = public.SealedBox(public_key)
    encrypted = sealed_box.encrypt(secret_value.encode('utf-8'))
    return b64encode(encrypted).decode('utf-8')


@typechecked
def get_repo_public_key(gh_api: GitHub, owner: str, repo_name: str) -> Dict[str, str]:
    """Get repository public key (salt) for encryption.

    Details: https://developer.github.com/v3/actions/secrets/#get-a-repository-public-key

    Args:
        gh_api: (agithub.GitHub.GitHub) Github API object
        owner: (str) Repository owner (user or organization).
        repo_name: (str) Repository name.

    Returns:
        Dict[str, str]: `{'key': str, 'key_id': str}` where:
            `key` is public key (salt) for encryption.
            `key_id` - Github salt ID.
    """
    for attempts_left in range(3, -1, -1):
        try:
            status, public_key = getattr(
                getattr(
                    getattr(
                        gh_api.repos, owner,
                    ), repo_name,
                ).actions.secrets, 'public-key',
            ).get()

        except TimeoutError:
            logger.warning(f"Can't access Github. Timeout error. Attempts left: {attempts_left}")
        else:
            break
    else:
        sys.exit("ERROR: Can't access Github. Please, try again later")

    if status != 200:  # noqa: WPS432
        sys.exit(f"Can't get repo public-key. Full error: {public_key}")

    return public_key


@typechecked
def add_secret(  # noqa: WPS211
    gh_api: GitHub,
    owner: str,
    repo_name: str,
    key: str,
    body: Dict[str, str],
):
    """Add/update Github encrypted KeyValue storage - Github repo Secrets.

    Details: https://developer.github.com/v3/actions/secrets/#create-or-update-a-repository-secret

    Args:
        gh_api: (agithub.GitHub.GitHub) Github API object
        owner: (str) Repository owner (user or organization).
        repo_name: (str) Repository name.
        key: (str) Secret name.
        body: (Dict[str, str]) Secret body, that include: `{'encrypted_value': str, 'key_id': str}`.
    """
    for attempts_left in range(3, -1, -1):
        try:
            status, response = getattr(
                getattr(
                    getattr(
                        gh_api.repos, owner,
                    ), repo_name,
                ).actions.secrets, key,
            ).put(
                body=body,
            )
        except TimeoutError:
            logger.warning(f"Can't access Github. Timeout error. Attempts left: {attempts_left}")
        else:
            break
    else:
        sys.exit("ERROR: Can't access Github. Please, try again later")

    if status not in {201, 204}:
        sys.exit(f'ERROR ocurred when try populate access_key to repo. {response}')


@typechecked
def create_secrets(  # noqa: WPS211
    creds: Dict[str, Union[str, bool]],
    cloud: str,
    owner: str,
    repo_name: str,
    git_token: str,
):
    """Create Github repo secrets for specified Cloud.

    Args:
        creds: (Dict[str, Union[str, bool]]) Programmatic access to cloud:
            `{ 'key': str, 'secret': str, ... }`.
        cloud: (str) Cloud name.
        owner: (str) Repository owner (user or organization).
        repo_name: (str) Repository name.
        git_token: (str) Git Provider token.
    """
    gh_api = GitHub(token=git_token)

    # Get public key for encryption
    public_key = get_repo_public_key(gh_api, owner, repo_name)

    if cloud == 'AWS':
        key = 'AWS_ACCESS_KEY_ID'
        secret = 'AWS_SECRET_ACCESS_KEY'  # noqa: S105

    # Put secret to repo
    body = {
        'encrypted_value': encrypt(public_key['key'], creds['key']),
        'key_id': public_key['key_id'],
    }
    add_secret(gh_api, owner, repo_name, key, body)

    body = {
        'encrypted_value': encrypt(public_key['key'], creds['secret']),
        'key_id': public_key['key_id'],
    }
    add_secret(gh_api, owner, repo_name, secret, body)

    logger.info('Secrets added to Github repo')


@typechecked
def get_last_release(owner: str = 'shalb', repo: str = 'cluster.dev') -> str:
    """Get last Github repo release.

    Args:
        owner: (str) repo owner name. Default: 'shalb'.
        repo: (str) repo name. Default: 'cluster.dev'.

    Returns:
        realise_version string.
    """
    gh_api = GitHub()
    status, response = getattr(
        getattr(
            gh_api.repos, owner,
        ), repo,
    ).releases.latest.get()

    if status != 200:  # noqa: WPS432
        sys.exit(f"Can't access Github. Full error: {response}")

    return response['tag_name']
