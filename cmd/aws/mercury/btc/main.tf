provider "aws" {
  version = "~> 2.0"
  region  = var.region
}

// *** Input variables ***
variable "btc_mainnet_username" {
  description = "jsonrpc username for bitcoin mainnet nodes"
}

variable "btc_mainnet_password" {
  description = "jsonrpc password for bitcoin mainnet nodes"
}

variable "btc_testnet_username" {
  description = "username for bitcoin testnet nodes jsonrpc"
}

variable "btc_testnet_password" {
  description = "password for bitcoin tesnet nodes jsonrpc"
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
#[core]
server=1
listen=1
txindex=1
dbcache=6000

#[rpc]
rpcbind=0.0.0.0
rpcuser=${var.btc_mainnet_username}
rpcpassword=${var.btc_mainnet_password}
rpcallowip=0.0.0.0/0
rpcthreads=6
EOF

  config_file_testnet = <<EOF
#[core]
dbcache=3000
testnet=1
server=1
listen=1
txindex=1
zmqpubrawblock=tcp://0.0.0.0:28332
zmqpubrawtx=tcp://0.0.0.0:28333

#[rpc]
rpcuser=${var.btc_testnet_username}
rpcpassword=${var.btc_testnet_password}
rpcallowip=0.0.0.0/0
rpcthreads=6
rpcworkqueue=128

[test]
rpcbind=0.0.0.0
EOF

  service_file = <<EOF
[Unit]
Description=Bitcoin's distributed currency daemon
After=network.target

[Service]
User=bitcoin
Group=bitcoin

Type=forking
ExecStart=/usr/bin/bitcoind -daemon

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
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-bionic-18.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}