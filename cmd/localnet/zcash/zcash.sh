zcashd
sleep 5

zcash-cli importaddress $Address 
zcash-cli generate 100

while :
do
    zcash-cli generate 1
    sleep 75
done