module "ec2" {
  source = "./mercury/btc"

  region = "us-west-2"
  key_name = "mercury-testing"
  vpc_id = "vpc-2fed7f49"
  private_key_file = "~/.ssh/mercury-testing.pem"
}