#!/bin/bash
Address=$1
SegWitAddress=$2
bitcoind
sleep 5

bitcoin-cli importaddress $Address 
bitcoin-cli generatetoaddress 120 $Address

if [ "$SegWitAddress" != "" ]
then
    bitcoin-cli importaddress $SegWitAddress 
    bitcoin-cli generatetoaddress 10 $SegWitAddress
fi   

while :
do
    bitcoin-cli generatetoaddress 1 $Address
    sleep 600
done