package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/renproject/mercury/types/ethtypes"
)

var (
	paramNetwork  string
	paramAddress  string
	paramName     string
	paramABI      string
	paramArtifact string
)

func main() {
	flag.StringVar(&paramAddress, "address", "", "Address of the contract")
	flag.StringVar(&paramName, "name", "", "Name of the contract, the newly generated interface is named after this")
	flag.StringVar(&paramNetwork, "network", "mainnet", "EVM chain network")
	flag.StringVar(&paramABI, "abi", "", "ABI of the contract")
	flag.StringVar(&paramArtifact, "artifact", "", "Provide truffle artifact to generate the bindings from")
	flag.Parse()

	if paramArtifact != "" {
		artifact, err := ioutil.ReadFile(paramArtifact)
		if err != nil {
			panic(fmt.Errorf("failed to read artifact from file: %v", paramArtifact))
		}
		obj := struct {
			ContractName string          `json:"contractName"`
			ABI          json.RawMessage `json:"abi"`
		}{}
		if err := json.Unmarshal(artifact, &obj); err != nil {
			panic(fmt.Errorf("%v is not a valid truffle artifact", paramArtifact))
		}
		if err := writeBindingsToFile(string(obj.ABI), obj.ContractName); err != nil {
			panic(err)
		}
		return
	}

	if paramName == "" {
		panic("please provide a name for this contract, the interface is named after this")
	}
	if paramABI != "" {
		if err := writeBindingsToFile(paramABI, paramName); err != nil {
			panic(err)
		}
		return
	}
	if paramAddress == "" {
		panic("please provide the address of contract, this is used to recover the ABI of the contract")
	}
	abiString, err := getContractDetails(paramNetwork, ethtypes.AddressFromHex(paramAddress))
	if err != nil {
		panic(err)
	}
	if err := writeBindingsToFile(abiString, paramName); err != nil {
		panic(err)
	}
}

func writeBindingsToFile(abi, name string) error {
	contractABI, err := newABI(abi)
	if err != nil {
		return err
	}
	d1 := []byte(buildImports(name))
	d2 := []byte(buildInterface(name, contractABI))
	d3 := []byte(buildConstructor(name, abi))
	d4 := []byte(buildFunctions(name, contractABI))
	if err := ioutil.WriteFile(fmt.Sprintf("./%s.go", name), append(append(d1, d2...), append(d3, d4...)...), 0644); err != nil {
		return err
	}
	return nil
}

func getContractDetails(network string, address ethtypes.Address) (string, error) {
	apiPrefix := ""
	switch strings.ToLower(network) {
	case "mainnet":
		apiPrefix = "api"
	case "kovan":
		apiPrefix = "api-kovan"
	case "ropsten":
		apiPrefix = "api-ropsten"
	default:
		return "", fmt.Errorf("unknown network: %s", network)
	}

	url := fmt.Sprintf("https://%s.etherscan.io/api?module=contract&action=getabi&address=%s&apikey=R8F2CVXTVSCIDD2IQ2ZQP9P6VZADUWHDHN", apiPrefix, address.Hex())
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	response := struct {
		ABI string `json:"result"`
	}{}
	if err := json.Unmarshal(respBytes, &response); err != nil {
		return "", err
	}
	return response.ABI, nil
}

func buildImports(contractName string) string {
	return fmt.Sprintf("package %s\n\nimport (\n\t\"context\"\n\t\"fmt\"\n\t\"math/big\"\n\t\"github.com/renproject/mercury/sdk/client/ethclient\"\n\t\"github.com/renproject/mercury/types/ethtypes\"\n)\n", contractName)
}

func buildConstructor(contractName, contractABI string) string {
	return fmt.Sprintf("type %s struct {\n\tcontract ethtypes.Contract\n}\n\nvar ABI = `%s`\n\n// New returns a new %s instance\nfunc New(client ethclient.Client, addr ethtypes.Address) (%s, error) {\n\tcontract, err := client.Contract(addr, []byte(ABI))\n\tif err != nil {\n\t\treturn nil, fmt.Errorf(\"failed to bind %s at address=%s\", addr, err)\n\t}\n\treturn &%s{\n\t\tcontract: contract,\n\t}, nil\n}\n", structName(contractName), contractABI, strings.Title(contractName), strings.Title(contractName), contractName, "%v: %v", structName(contractName))
}

func buildFunctions(contractName string, contractABI contractABI) string {
	funcString := ""
	for _, method := range contractABI {
		if method.Type == "function" {
			funcString = fmt.Sprintf("%sfunc (c *%s) %s%s\n\n", funcString, structName(contractName), functionSig(method), functionBody(method))
		}
	}
	return funcString
}

func buildInterface(contractName string, contractABI contractABI) string {
	interfaceString := fmt.Sprintf("type %s interface{\n", strings.Title(contractName))
	for _, method := range contractABI {
		if method.Type == "function" {
			interfaceString = fmt.Sprintf("%s\t%s\n", interfaceString, functionSig(method))
		}
	}
	return fmt.Sprintf("%s}\n\n", interfaceString)
}

func functionSig(method method) string {
	return fmt.Sprintf("%s(ctx context.Context%s) (%serror)", strings.Title(method.Name), formatInArgs(method.Constant, method.Payable, method.Inputs), formatOutArgs(method.Constant, method.Outputs))
}

func functionBody(method method) string {
	params := ""
	for _, in := range method.Inputs {
		if in.Name[:1] == "_" {
			in.Name = in.Name[1:]
		}
		params = fmt.Sprintf("%s, %s", params, in.Name)
	}
	if method.Payable {
		return fmt.Sprintf("{\n\treturn c.contract.BuildTx(ctx, signer, \"%s\", value%s)\n}", method.Name, params)
	}
	if !method.Constant {
		return fmt.Sprintf("{\n\treturn c.contract.BuildTx(ctx, signer, \"%s\", nil%s)\n}", method.Name, params)
	}
	declaration := fmt.Sprintf("\n\targ := new(%s)", method.Outputs[0].GoType)
	return fmt.Sprintf("{%s\n\tif err := c.contract.Call(ctx, ethtypes.Address{}, arg, \"%s\"%s); err != nil{\n\t\treturn %s\n\t}\n\treturn *arg, nil\n}", declaration, method.Name, params, errVals(method.Outputs))
}

func formatInArgs(constant, payable bool, inArgs []arg) string {
	var formattedArgs string
	for _, inArg := range inArgs {
		if inArg.Name[:1] == "_" {
			inArg.Name = inArg.Name[1:]
		}
		formattedArgs = fmt.Sprintf("%s, %s %s", formattedArgs, inArg.Name, inArg.GoType)
	}
	if !constant {
		formattedArgs = fmt.Sprintf("%s, signer ethtypes.Address", formattedArgs)
	}
	if payable {
		formattedArgs = fmt.Sprintf("%s, value *big.Int", formattedArgs)
	}
	return formattedArgs
}

func formatOutArgs(constant bool, outArgs []arg) string {
	if !constant {
		return "ethtypes.Tx, "
	}
	var formattedArgs string
	for _, outArg := range outArgs {
		formattedArgs = fmt.Sprintf("%s%s, ", formattedArgs, outArg.GoType)
	}
	return formattedArgs
}

func formatType(argType abi.Type) string {
	typeString := argType.Type.String()
	if typeString == "common.Address" {
		return "ethtypes.Address"
	}
	return typeString
}

func errVals(args []arg) string {
	errMsg := "return "
	for _, arg := range args {
		errMsg = fmt.Sprintf("%s, ", defaultValue(arg.GoType))
	}
	return fmt.Sprintf("%serr", errMsg)
}

func defaultValue(goType string) string {
	if goType[:4] == "uint" || goType[:3] == "int" {
		return "0"
	}
	if len(goType) >= 6 && (goType[len(goType)-5:] == "]byte" || goType[6:] == "struct") {
		return fmt.Sprintf("%s{}", goType)
	}
	switch goType {
	case "bool":
		return "false"
	case "string":
		return "\"\""
	case "*big.Int":
		return "nil"
	case "ethtypes.Address":
		return "ethtypes.Address{}"
	default:
		panic(fmt.Sprintf("failed to get default value of type: %s", goType))
	}
}

type method struct {
	Name            string `json:"name"`
	Inputs          []arg  `json:"inputs"`
	Type            string `json:"type"`
	Outputs         []arg  `json:"outputs,omitempty"`
	Payable         bool   `json:"payable,omitempty"`
	Constant        bool   `json:"constant,omitempty"`
	Anonymous       bool   `json:"anonymus,omitempty"`
	StateMutability string `json:"stateMutability,omitempty"`
}

type arg struct {
	Indexed bool   `json:"bool,omitempty"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	GoType  string
}

type contractABI []method

func newABI(abiString string) (contractABI, error) {
	ethABI, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		return nil, err
	}
	cABI := contractABI{}
	json.Unmarshal([]byte(abiString), &cABI)
	for i, method := range cABI {
		for j, in := range ethABI.Methods[method.Name].Inputs {
			cABI[i].Inputs[j].GoType = formatType(in.Type)
		}
		for j, out := range ethABI.Methods[method.Name].Outputs {
			cABI[i].Outputs[j].GoType = formatType(out.Type)
		}
	}
	return cABI, nil
}

func structName(contractName string) string {
	if contractName[0] >= 65 && contractName[0] <= 90 {
		return string(append([]byte{contractName[0] + byte(0x20)}, contractName[1:]...))
	}
	return contractName
}
