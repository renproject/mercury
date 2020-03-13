provider "aws" {
  region = var.region
}

// Retrive the ubuntu ami from the marketplace
data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name = "name"
    values = [
      "ubuntu/images/hvm-ssd/ubuntu-bionic-18.04-amd64-server-*"]
  }

  filter {
    name = "virtualization-type"
    values = [
      "hvm"]
  }

  owners = [
    "099720109477"]
  # Canonical
}

module "bitcoin" {
  source = "./btc"

  btc_mainnet_username = var.btc_mainnet_username
  btc_mainnet_password = var.btc_mainnet_password
  btc_testnet_username = var.btc_testnet_username
  btc_testnet_password = var.btc_testnet_password

  region = var.region
  available_zone_1 = var.avaiable_zone_1
  available_zone_2 = var.avaiable_zone_2

  vpc_id = aws_vpc.aws_vpc_mercury.id
  subnet_id_1 = aws_subnet.aws_subnet1.id
  subnet_id_2 = aws_subnet.aws_subnet2.id

  key_name = var.key_name
  private_key_file = var.key_file
}

//module "zcash" {
//  source = "./zec"
//
//  zec_mainnet_username = var.zec_mainnet_username
//  zec_mainnet_password = var.zec_mainnet_password
//  zec_testnet_username = var.zec_testnet_username
//  zec_testnet_password = var.zec_testnet_password
//
//  region = var.region
//  available_zone_1 = var.avaiable_zone_1
//  available_zone_2 = var.avaiable_zone_2
//
//  vpc_id = aws_vpc.aws_vpc_mercury.id
//  subnet_id_1 = aws_subnet.aws_subnet1.id
//  subnet_id_2 = aws_subnet.aws_subnet2.id
//
//  key_name = var.key_name
//  private_key_file = var.key_file
//}