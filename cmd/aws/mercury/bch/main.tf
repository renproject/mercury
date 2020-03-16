provider "aws" {
  version = "~> 2.0"
  region = var.region
}

// *** Input variables ***
// Check README.md file to see the description of each input parameter.
variable "bch_mainnet_username" {}

variable "bch_mainnet_password" {}

variable "bch_testnet_username" {}

variable "bch_testnet_password" {}

variable "region" {}

variable "available_zone_1" {}

variable "available_zone_2" {}

variable "key_name" {}

variable "key_file" {}

variable "ami_id" {}

variable "default_sg_id" {}

variable "vpc_id" {}

variable "subnet_id_1" {}

variable "subnet_id_2" {}

// *** Local variables ***
locals {
  config_file_mainnet = <<EOF
rpcuser=${var.bch_mainnet_username}
rpcpassword=${var.bch_mainnet_password}
server=1
listen=0
txindex=1
dbcache=3000
rpcallowip=0.0.0.0/0
rpcthreads=6
EOF

  config_file_testnet = <<EOF
rpcuser=${var.bch_testnet_username}
rpcpassword=${var.bch_testnet_password}
testnet=1
server=1
listen=0
txindex=1
rpcallowip=0.0.0.0/0
rpcworkqueue=128
dbcache=3000
rpcthreads=6

[test]
rpcbind=0.0.0.0
EOF

  service_file = <<EOF
[Unit]
Description=Bitcoin Cash's distributed currency daemon
After=network.target

[Service]
User=bitcoin
Group=bitcoin

Type=forking
ExecStart=/usr/bin/bitcoind --daemon
ExecStop=/usr/bin/bitcoin-cli stop

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