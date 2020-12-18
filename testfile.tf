
data "terraform_remote_state" "infra-dev-vpc" {
  backend = "s3"
  config = {
    bucket = "new-cluster-dev"
    key    = "infra-dev/vpc"
    region = "eu-central-1"
  }
  asdasd = {
    bucket = "new-cluster-dev"
    key    = "infra-dev/vpc"
    region = "replacer.${test.var}sdsd.sdsdd"
  }
}
