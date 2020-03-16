The mercury module will deploy mercury instance and all blockchain nodes.
## Usage ##

1. Make sure you have the `credentials` file in your `~/.aws` folder with your aws credentials.
2. Create a ssh key from aws panel and remember the name of the key. The key needs to be in the same
   region you want to deploy.
2. Create `.tfvars` file in this folder which has all required input variables.
   You should be able to find one in 1password.   
3. Run `terraform init` in this folder and you should see no errors.
4. Run `terraform apply -var-file=.tfvars --auto-approve` which will takes some time to finish. 

## Inputs ##

- `region`           : region on aws where to deploy the mercury infrastructure. i.e. `us-west-2`
- `available_zone_1` : first available zone we want to use in the region. i.e. `us-west-2a`
- `available_zone_2` : second available zone we want to use in the region. i.e. `us-west-2b`

- `key_name`         : name of the ssh key pair   
- `private_key_file` : file path of the ssh key file 

- `btc_mainnet_username` : jsonrpc username for bitcoin mainnet nodes 
- `btc_mainnet_password` : jsonrpc password for bitcoin mainnet nodes
- `btc_testnet_username` : jsonrpc username for bitcoin testnet nodes 
- `btc_testnet_password` : jsonrpc password for bitcoin testnet nodes
- `zec_mainnet_username` : jsonrpc username for zcash mainnet nodes 
- `zec_mainnet_password` : jsonrpc password for zcash mainnet nodes
- `zec_testnet_username` : jsonrpc username for zcash testnet nodes 
- `zec_testnet_password` : jsonrpc password for zcash testnet nodes
- `bch_mainnet_username` : jsonrpc username for bitcoin cash mainnet nodes 
- `bch_mainnet_password` : jsonrpc password for bitcoin cash mainnet nodes
- `bch_testnet_username` : jsonrpc username for bitcoin cash testnet nodes 
- `bch_testnet_password` : jsonrpc password for bitcoin cash testnet nodes

- `infura_key_darknode` : infura id of project darknode
- `infura_key_dcc`      : infura id of project dcc
- `infura_key_default`  : infura id of project default  
- `infura_key_renex`    : infura id of project renex  
- `infura_key_renex_ui` : infura id of project renex_ui  
- `infura_key_swapperd` : infura id of project swapperd  

## Outputs ## 

- `mercury_load_balancer_dns` : dns name of the mercury load-balancer

