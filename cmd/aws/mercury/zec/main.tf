provider "aws" {
  version = "~> 2.0"
  region = var.region
}

// *** Input variables ***
// Check README.md file to see the description of each input parameter.
variable "zec_mainnet_username" {}

variable "zec_mainnet_password" {}

variable "zec_testnet_username" {}

variable "zec_testnet_password" {}

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