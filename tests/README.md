To run all action local for tests:
1) Edit tests.sh, set variables USER and PASS (aws_access_key_id and aws_secret_access_key)
2) cd tests/ && ./tests.sh

Destroy infrastructure, created by terraform:

cd ../terraform/aws/(module) && terraform destroy
(then press enter for leave variables empty, set only region)
