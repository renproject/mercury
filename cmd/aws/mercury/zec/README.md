The zec module will deploy all zcash-related services in Mercury to AWS, including :  

- Zcash mainnet nodes * 2 in different available zones. 
- Zcash testnet nodes * 1 in `available_zone_1`
- A load-balancer in front of the mainnet nodes.

## Inputs ##

- `zec_mainnet_username` : jsonrpc username for zcash mainnet nodes 
- `zec_mainnet_password` : jsonrpc password for zcash mainnet nodes
- `zec_testnet_username` : jsonrpc username for zcash testnet nodes 
- `zec_testnet_password` : jsonrpc password for zcash testnet nodes

- `region`           : region on aws where to deploy the mercury infrastructure. i.e. `us-west-2`
- `available_zone_1` : first available zone we want to use in the region. i.e. `us-west-2a`
- `available_zone_2` : second available zone we want to use in the region. i.e. `us-west-2b`
 
- `key_name`         : name of the ssh key pair   
- `private_key_file` : file path of the ssh key file  

- `ami_id`           : ami id of the ubuntu image
- `default_sg_id`    : id of the default security group 
- `vpc_id`           : id of the vpc in the region 
- `subnet_id_1`      : id of the subnet in the first available zone
- `subnet_id_2`      : id of the subnet in the first available zone

## Outputs ##

- `zec_lb_dns`     : DNS name of the zcash load balancer.
- `zec_testnet_ip` : testnet node private ip address.