The zec module will deploy all zcash-related services in Mercury to AWS. 
It expects the following inputs :

### jsonrpc credentials 
- `zec_mainnet_username` : jsonrpc username for zcash mainnet nodes 
- `zec_mainnet_password` : jsonrpc password for zcash mainnet nodes
- `zec_testnet_username` : jsonrpc username for zcash testnet nodes 
- `zec_testnet_password` : jsonrpc password for zcash testnet nodes

### Region and available zones  
- `region`           : region on aws where to deploy the infrastructure
- `available_zone_1` : first available zone we want to use in the region 
- `available_zone_2` : second available zone we want to use in the region
 
### VPC details  
- `vpc_id`           : id of the vpc in the region. 
- `subnet_id_1`      : id of the subnet in the first available zone
- `subnet_id_2`      : id of the subnet in the first available zone

### SSH key 
- `key_name`         : name of the ssh key pair   
- `private_key_file` : file path of the private file which can ssh into the nodes  

And will deploy the following services:

- Zcash mainnet nodes * 2 
  - Security Group 
  - EC2 instance * 2 in different available zones.
- Zcash testnet nodes * 1 
  - Security Group 
  - EC2 instance * 1 in a random available zone.
- A load-balancer in front of the mainnet nodes.
  - ELB * 1 
