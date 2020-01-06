touch /tmp/test-file.txt
echo "test" > /tmp/test-file.txt
pwd

# Install AWS utils
yum install -y wget unzip
cd /usr/local/bin
wget http://s3.amazonaws.com/ec2-downloads/ec2-api-tools.zip
unzip ec2-api-tools.zip
mv ec2-api-tools-* /usr/local/bin/ec2-api-tools

# execute listing for s3
aws s3 ls > /root/aws-s3-output.txt