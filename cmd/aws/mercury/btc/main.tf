provider "aws" {
  version = "~> 2.0"
  region = var.region
}

// *** Input variables ***
// Check README.md file to see the description of each input parameter.
variable "btc_mainnet_username" {}

variable "btc_mainnet_password" {}

variable "btc_testnet_username" {}

variable "btc_testnet_password" {}

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