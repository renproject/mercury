provider "aws" {
  version = "~> 2.0"
  region = var.region
}

// *** Input variables ***
variable "zec_mainnet_username" {
  description = "jsonrpc username for zcash mainnet nodes"
}

variable "zec_mainnet_password" {
  description = "jsonrpc password for zcash mainnet nodes"
}

variable "zec_testnet_username" {
  description = "username for zcash testnet nodes jsonrpc"
}

variable "zec_testnet_password" {
  description = "password for zcash tesnet nodes jsonrpc"
}

variable "region" {
  description = "region on aws where to deploy the infrastructure"
}

variable "available_zone_1" {
  description = "first available zone we want to use"
}

variable "available_zone_2" {
  description = "second available zone we want to use"
}

variable "vpc_id" {
  description = "id of the vpc in the region"
}

variable "subnet_id_1" {
  description = "id of the subnet in the first available zone"
}

variable "subnet_id_2" {
  description = "id of the subnet in the second available zone"
}

variable "key_name" {
  description = "name of the ssh key pair"
}

variable "private_key_file" {
  description = "file path of the private key"
}

// *** Local variables ***
locals {
  config_file_mainnet = <<EOF
addnode=mainnet.z.cash
rpcuser=${var.zec_mainnet_username}
rpcpassword=${var.zec_mainnet_password}
server=1
listen=0
txindex=1
dbcache=3000
rpcallowip=0.0.0.0/0
EOF

  config_file_testnet = <<EOF
addnode=testnet.z.cash
rpcuser=${var.zec_testnet_username}
rpcpassword=${var.zec_testnet_password}
testnet=1
server=1
listen=0
txindex=1
rpcallowip=0.0.0.0/0
rpcworkqueue=128
dbcache=3000

[test]
rpcbind=0.0.0.0
EOF

  service_file = <<EOF
[Unit]
Description=Zcash's distributed currency daemon
After=network.target

[Service]
User=zcash
Group=zcash

Type=forking
ExecStart=/usr/bin/zcashd --daemon
ExecStop=/usr/bin/zcash-cli stop

Restart=always
PrivateTmp=true
TimeoutStopSec=60s
TimeoutStartSec=2s
StartLimitInterval=120s
StartLimitBurst=5

[Install]
WantedBy=multi-user.target
EOF
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