#!/bin/bash
Address=$1
bitcoind
sleep 5

bitcoin-cli importaddress $Address 
bitcoin-cli generatetoaddress 120 $Address

while :
do
    bitcoin-cli generatetoaddress 1 $Address
    sleep 600
done