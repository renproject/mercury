#!/bin/bash

Address=$1
echo "
mineraddress=$Address" >> ~/.zcash/zcash.conf
zcashd
sleep 5

zcash-cli importaddress $Address 
zcash-cli generate 120

while :
do
    zcash-cli generate 1
    sleep 75
done