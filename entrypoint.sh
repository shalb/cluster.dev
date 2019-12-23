#!/bin/sh -l

echo "Hello $1"
time=$(date)
find .
echo ::set-output name=time::$time