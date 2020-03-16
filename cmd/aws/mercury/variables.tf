// Check READ.md to see the description of each variable

// ** Region and available zones **
variable "region" {}

variable "avaiable_zone_1" {}

variable "avaiable_zone_2" {}

// ** SSH key **
variable "key_name" {}

variable "key_file" {}

// ** JSONRPC usernames and passwords for all blockchains **
variable "btc_mainnet_username" {}

variable "btc_mainnet_password" {}

variable "btc_testnet_username" {}

variable "btc_testnet_password" {}

variable "zec_mainnet_username" {}

variable "zec_mainnet_password" {}

variable "zec_testnet_username" {}

variable "zec_testnet_password" {}

variable "bch_mainnet_username" {}

variable "bch_mainnet_password" {}

variable "bch_testnet_username" {}

variable "bch_testnet_password" {}

// ** Infura Project ID **
variable "infura_key_darknode" {}

variable "infura_key_dcc" {}

variable "infura_key_default" {}

variable "infura_key_renex" {}

variable "infura_key_renex_ui" {}

variable "infura_key_swapperd" {}

// ** Local variables **
locals {
  service_file = <<EOF
[Unit]
Description=Mercury Server

[Service]
WorkingDirectory=/home/mercury
ExecStart=/home/mercury/.mercury/bin/mercury
Restart=on-failure
PrivateTmp=true
NoNewPrivileges=true
EnvironmentFile=/home/mercury/mercury.env

# Specifies which signal to use when killing a service. Defaults to SIGTERM.
# SIGHUP gives parity time to exit cleanly before SIGKILL (default 90s)
KillSignal=SIGHUP

[Install]
WantedBy=default.target
EOF

  env_file = <<EOF
BITCOIN_MAINNET_RPC_URL=http://${module.bitcoin.btc_lb_dns}
BITCOIN_MAINNET_RPC_USERNAME=${var.btc_mainnet_username}
BITCOIN_MAINNET_RPC_PASSWORD=${var.btc_mainnet_password}
BITCOIN_TESTNET_RPC_URL=http://${module.bitcoin.btc_testnet_ip}:18332
BITCOIN_TESTNET_RPC_USERNAME=${var.btc_testnet_username}
BITCOIN_TESTNET_RPC_PASSWORD=${var.btc_testnet_password}

ZCASH_MAINNET_RPC_URL=http://${module.zcash.zec_lb_dns}
ZCASH_MAINNET_RPC_USERNAME=${var.zec_mainnet_username}
ZCASH_MAINNET_RPC_PASSWORD=${var.zec_mainnet_password}
ZCASH_TESTNET_RPC_URL=http://${module.zcash.zec_testnet_ip}:18232
ZCASH_TESTNET_RPC_USERNAME=${var.zec_testnet_username}
ZCASH_TESTNET_RPC_PASSWORD=${var.zec_testnet_password}

BCASH_MAINNET_RPC_URL=http://${module.bcash.bch_lb_dns}
BCASH_MAINNET_RPC_USERNAME=${var.bch_mainnet_username}
BCASH_MAINNET_RPC_PASSWORD=${var.bch_mainnet_password}
BCASH_TESTNET_RPC_URL=http://${module.bcash.bch_testnet_ip}:18332
BCASH_TESTNET_RPC_USERNAME=${var.bch_testnet_username}
BCASH_TESTNET_RPC_PASSWORD=${var.bch_testnet_password}

INFURA_KEY_DARKNODE=${var.infura_key_darknode}
INFURA_KEY_DCC=${var.infura_key_dcc}
INFURA_KEY_DEFAULT=${var.infura_key_default}
INFURA_KEY_RENEX=${var.infura_key_renex}
INFURA_KEY_RENEX_UI=${var.infura_key_renex_ui}
INFURA_KEY_SWAPPERD=${var.infura_key_swapperd}
EOF
}