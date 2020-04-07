#!/usr/bin/env python

import os
import sys
import json
import pprint
import argparse

# Use this script to parse AWS Cloud Trail logs

parser = argparse.ArgumentParser()
parser.add_argument('--logs_dir', default='./aws_logs/', help='Directory with unzipped AWS Cloud Trail logs in json format')
parser.add_argument('--ip_address', help='IP address of machine, which activity you want to cach. Please use separate host or awoid AWS Console access during logs collection if running both on the same PC')
parser.add_argument('--debug', default=False, help='enable debug mode')
args = parser.parse_args()

for root, subFolder, files in os.walk(args.logs_dir):
    for item in files:
        json_file_name = os.path.join(root,item)
        json_file=open(json_file_name)
        logs = json.load(json_file)
        for record in logs['Records']:
            try:
                if 'my-cloud-logs' in str(record):
                    continue
                if record['sourceIPAddress'] != args.ip_address:
                    continue
                print('{} | {} | {}'.format(record['eventSource'], record['eventName'], record['requestParameters']))
            except:
                pprint.pprint(record)

