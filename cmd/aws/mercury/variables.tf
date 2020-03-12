variable "region" {
  description = "region on aws where to deploy the infrastructure"
  default = "us-west-2"
}

variable "avaiable_zone_1" {
  description = "the first available zone we want to use."
  default = "us-west-2a"
}

variable "avaiable_zone_2" {
  description = "the first available zone we want to use."
  default = "us-west-2b"
}

variable "btc_mainnet_username" {
}

variable "btc_mainnet_password" {
}

variable "btc_testnet_username" {
}

variable "btc_testnet_password" {
}

variable "zec_mainnet_username" {
}

variable "zec_mainnet_password" {
}

variable "zec_testnet_username" {
}

variable "zec_testnet_password" {
}

variable "bch_mainnet_username" {
}

variable "bch_mainnet_password" {
}

variable "bch_testnet_username" {
}

variable "bch_testnet_password" {
}
