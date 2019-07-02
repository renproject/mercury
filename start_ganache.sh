#!/usr/bin/env bash

ganache-cli -p $GANACHE_PORT --account="$LOCAL_ETH_TESTNET_PRIVATE_KEY,100000000000000000000"
