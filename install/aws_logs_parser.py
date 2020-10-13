#!/usr/bin/env python
"""Use this script to parse AWS Cloud Trail logs."""
import argparse
import json
import os
import pprint

parser = argparse.ArgumentParser()
parser.add_argument(
    '--logs_dir',
    default='./aws_logs/',
    help='Directory with unzipped AWS Cloud Trail logs in json format',
)
parser.add_argument(
    '--ip_address',
    help='IP address of machine, which activity you want to cach. ' +
    'Please use separate host or awoid AWS Console access during ' +
    'logs collection if running both on the same PC',
)
parser.add_argument('--debug', default=False, help='enable debug mode')
args = parser.parse_args()

for root, _sub_folder, files in os.walk(args.logs_dir):
    for filename in files:
        json_file_name = os.path.join(root, filename)

        with open(json_file_name) as json_file:
            logs = json.load(json_file)

        for record in logs['Records']:
            try:  # noqa: WPS229 # FIXME
                if 'my-cloud-logs' in str(record):
                    continue
                if record['sourceIPAddress'] != args.ip_address:
                    continue
                pprint.pprint(
                    '{} | {} | {}'.format(  # noqa: P101 # FIXME
                        record['eventSource'],
                        record['eventName'],
                        record['requestParameters'],
                    ),
                )
            except:  # pylint: disable=bare-except # noqa: E722, B001 # FIXME
                pprint.pprint(record)
