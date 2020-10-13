# AWS IAM permissions

This document explains how to create or update [aws_policy.json](../install/installer/aws_policy.json).

* [Requirements](#requirements)
* [Collect logs](#collect-logs)
  * [Run installation and destroy](#run-installation-and-destroy)
  * [Copy logs](#copy-logs)
* [Parse logs](#parse-logs)
  * [Copy logs parsing script](#copy-logs-parsing-script)
  * [Get API calls with service](#get-api-calls-with-service)
  * [Get API calls with service and request](#get-api-calls-with-service-and-request)
* [Create policy](#create-policy)

## Requirements

1. Create admin account in AWS with `Access Key`. See the [AWS documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html).
2. Install `aws-cli`. See [AWS documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html).
3. Configure `aws-cli`

    ```bash
    aws configure
    ```

5. Create S3 bucket

    ```bash
    aws s3 mb s3://my-cloud-logs
    ```

5. Configure Cloud Trail.

    Any region is OK. See [AWS documentation](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-tutorial.html#tutorial-step2).

    Needed options:

    ```bash
    Name: any (example: my-cloud-logs)
    Trail settings:
        Apply trail to all regions: Yes
    Management events:
        Read/Write events All
    Data events:
        Select all S3 buckets in your account: Read, Write
    Storage location:
        S3 bucket: my-cloud-logs
    ```

## Collect logs

### Run installation and destroy

See [README.md](../README.md).

### Copy logs

Replace `MY-...` by your account, region, date.

Copy the logs from region `us-east-1`, because global API calls are logged in this region.

```bash
mkdir ./aws_logs/
aws s3 sync s3://my-cloud-logs/AWSLogs/MY-ACCOUNT-ID/CloudTrail/MY-REGION/MY-YEAR/MY-MONTH/MY-DAY/ ./aws_logs/
aws s3 sync s3://my-cloud-logs/AWSLogs/MY-ACCOUNT-ID/CloudTrail/us-east-1/MY-YEAR/MY-MONTH/MY-DAY/ ./aws_logs/
gzip -d ./aws_logs/*.gz
```

## Parse logs

### Copy logs parsing script

```bash
curl https://raw.githubusercontent.com/shalb/cluster.dev/master/install/aws_logs_parser.py > aws_logs_parser.py
```

### Get API calls with service

Replace `MY-IP` by your IP address, which is used to deploy the cluster.

```bash
./aws_logs_parser.py --ip_address=MY-IP | awk -F "|" '{print $1 $2}' | sort -u | less -Ni
```

### Get API calls with service and request

Replace `MY-IP` by your IP address, which is used to deploy the cluster.

```bash
./aws_logs_parser.py --ip_address=MY-IP | sort -u | less -Ni
```

## Create policy

1. Open visual [policy editor](https://console.aws.amazon.com/iam/home?#/policies$new?step=edit) and add needed permissions regarding the output of the [script](../install/aws_logs_parser.py).
2. Save new policy time to time if it has many records, to prevent results from being lost.
3. Check out its JSON version and save it to [aws_policy.json](../install/installer/aws_policy.json).
4. Push JSON version to repo.
