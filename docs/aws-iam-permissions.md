   * [Info](#info)
   * [Requirements](#requirements)
      * [Create admin account in aws with access key](#create-admin-account-in-aws-with-access-key)
      * [Install aws cli](#install-aws-cli)
      * [Configure aws cli](#configure-aws-cli)
      * [Create S3 bucket](#create-s3-bucket)
      * [Configure Cloud Trail](#configure-cloud-trail)
   * [Collect logs](#collect-logs)
      * [Run installation and destroy](#run-installation-and-destroy)
      * [Copy logs](#copy-logs)
   * [Parse logs](#parse-logs)
      * [Copy logs parsing script](#copy-logs-parsing-script)
      * [Get API calls with service](#get-api-calls-with-service)
      * [Get API calls with service and request](#get-api-calls-with-service-and-request)
   * [Create policy](#create-policy)

# Info

This document explains how to create or update [aws_policy.json](../install/aws_policy.json)

# Requirements

## Create admin account in aws with access key

See [aws documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html)

## Install aws cli

See [aws documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html)

## Configure aws cli

~~~~
aws configure
~~~~

## Create S3 bucket

~~~~
aws s3 mb s3://my-cloud-logs
~~~~

## Configure Cloud Trail

Any region is ok

See [aws documentation](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-tutorial.html#tutorial-step2)

Needed options:

~~~~
Name: any (example: my-cloud-logs)
Trail settings:
    Apply trail to all regions: Yes
Management events:
    Read/Write events All
Data events:
    Select all S3 buckets in your account: Read, Write 
Storage location:
    S3 bucket: my-cloud-logs
~~~~

# Collect logs

## Run installation and destroy

See [readme](../README.md)

## Copy logs

Replace 'MY-...' by your account, region, date

Copy logs from region 'us-east-1', because global API calls logged in this region.

~~~~
mkdir ./aws_logs/
aws s3 sync s3://my-cloud-logs/AWSLogs/MY-ACCOUNT-ID/CloudTrail/MY-REGION/MY-YEAR/MY-MONTH/MY-DAY/ ./aws_logs/
aws s3 sync s3://my-cloud-logs/AWSLogs/MY-ACCOUNT-ID/CloudTrail/us-east-1/MY-YEAR/MY-MONTH/MY-DAY/ ./aws_logs/
gzip -d ./aws_logs/*.gz
~~~~

# Parse logs

## Copy logs parsing script

~~~~
curl https://raw.githubusercontent.com/shalb/cluster.dev/master/install/aws_logs_parser.py > aws_logs_parser.py
~~~~

## Get API calls with service

Replace MY-IP by your IP address, which used to deploy cluster

~~~~
./aws_logs_parser.py --ip_address=MY-IP | awk -F "|" '{print $1 $2}' | sort -u | less -Ni
~~~~

## Get API calls with service and request

Replace MY-IP by your IP address, which used to deploy cluster

~~~~
./aws_logs_parser.py --ip_address=MY-IP | sort -u | less -Ni
~~~~

# Create policy

Open visual [policy editor](https://console.aws.amazon.com/iam/home?#/policies$new?step=edit) and add needed permissions regarding to output of the [script](../install/aws_logs_parser.py)

Save new policy time to time if it has many records, to prevent results loss.

Check out its JSON version and save it to [aws_policy.json](../install/aws_policy.json)

Push JSON version to repo.

