provider "aws" {
  region = var.region
}

//module "bitcoin" {
//  source = "./btc"
//
//  btc_mainnet_username = var.btc_mainnet_username
//  btc_mainnet_password = var.btc_mainnet_password
//  btc_testnet_username = var.btc_testnet_username
//  btc_testnet_password = var.btc_testnet_password
//
//  region = var.region
//  available_zone_1 = var.avaiable_zone_1
//  available_zone_2 = var.avaiable_zone_2
//
//  vpc_id = aws_vpc.aws_vpc_mercury.id
//  subnet_id_1 = aws_subnet.aws_subnet1.id
//  subnet_id_2 = aws_subnet.aws_subnet2.id
//
//  key_name = "mercury-testing"
//  private_key_file = "~/.ssh/mercury-testing.pem"
//}

module "zcash" {
  source = "./zec"

  zec_mainnet_username = var.zec_mainnet_username
  zec_mainnet_password = var.zec_mainnet_password
  zec_testnet_username = var.zec_testnet_username
  zec_testnet_password = var.zec_testnet_password

  region = var.region
  available_zone_1 = var.avaiable_zone_1
  available_zone_2 = var.avaiable_zone_2

  vpc_id = aws_vpc.aws_vpc_mercury.id
  subnet_id_1 = aws_subnet.aws_subnet1.id
  subnet_id_2 = aws_subnet.aws_subnet2.id

  key_name = "mercury-testing"
  private_key_file = "~/.ssh/mercury-testing.pem"
}