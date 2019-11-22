package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/renproject/mercury/types/ethtypes"
)

var (
	paramNetwork  string
	paramAddress  string
	paramName     string
	paramABI      string
	paramBIN      string
	paramArtifact string
)

func main() {
	flag.StringVar(&paramAddress, "address", "", "Address of the contract")
	flag.StringVar(&paramName, "name", "", "Name of the contract, the newly generated interface is named after this")
	flag.StringVar(&paramNetwork, "network", "mainnet", "EVM chain network")
	flag.StringVar(&paramABI, "abi", "", "ABI of the contract")
	flag.StringVar(&paramBIN, "bin", "", "BIN of the contract")
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
			BIN          string          `json:"bytecode"`
		}{}
		if err := json.Unmarshal(artifact, &obj); err != nil {
			panic(fmt.Errorf("%v is not a valid truffle artifact", paramArtifact))
		}
		if err := writeBindingsToFile(string(obj.ABI), obj.BIN, obj.ContractName); err != nil {
			panic(err)
		}
		return
	}

	if paramName == "" {
		panic("please provide a name for this contract, the interface is named after this")
	}
	if paramABI != "" {
		if paramBIN == "" {
			panic("please provide the binary of the contract")
		}
		if err := writeBindingsToFile(paramABI, paramBIN, paramName); err != nil {
			panic(err)
		}
		return
	}
	if paramAddress == "" {
		panic("please provide the address of contract, this is used to recover the ABI of the contract")
	}
	abiString, binString, err := getContractDetails(paramNetwork, ethtypes.AddressFromHex(paramAddress))
	if err != nil {
		panic(err)
	}
	if err := writeBindingsToFile(abiString, binString, paramName); err != nil {
		panic(err)
	}
}

func writeBindingsToFile(abi, bin, name string) error {
	contractABI, constructorMethod, err := newABI(abi)
	if err != nil {
		return err
	}

	d1 := []byte(buildImports(name))
	d2 := []byte(buildInterface(name, contractABI))
	d3 := []byte(buildConstructor(name, abi, bin, constructorMethod))
	d4 := []byte(buildFunctions(name, contractABI))
	if err := ioutil.WriteFile(fmt.Sprintf("./%s.go", name), append(append(d1, d2...), append(d3, d4...)...), 0644); err != nil {
		return err
	}
	return nil
}

func getContractDetails(network string, address ethtypes.Address) (string, string, error) {
	apiPrefix := ""
	switch strings.ToLower(network) {
	case "mainnet":
		apiPrefix = "api"
	case "kovan":
		apiPrefix = "api-kovan"
	case "ropsten":
		apiPrefix = "api-ropsten"
	default:
		return "", "", fmt.Errorf("unknown network: %s", network)
	}

	binString := getBin(address, network)
	url := fmt.Sprintf("https://%s.etherscan.io/api?module=contract&action=getabi&address=%s&apikey=R8F2CVXTVSCIDD2IQ2ZQP9P6VZADUWHDHN", apiPrefix, address.Hex())
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	response := struct {
		ABI string `json:"result"`
	}{}
	if err := json.Unmarshal(respBytes, &response); err != nil {
		return "", "", err
	}
	return response.ABI, binString, nil
}

func buildImports(contractName string) string {
	return fmt.Sprintf("package %s\n\nimport (\n\t\"context\"\n\t\"encoding/hex\"\n\t\"fmt\"\n\t\"math/big\"\n\t\"github.com/renproject/mercury/sdk/client/ethclient\"\n\t\"github.com/renproject/mercury/types/ethtypes\"\n)\n", contractName)
}

func buildConstructor(contractName, contractABI, contractBIN string, constructorSig method) string {
	return fmt.Sprintf("type %s struct {\n\tcontract ethtypes.Contract\n}\n\nvar ABI = `%s`\n\nvar BIN = `%s`\n\n// New returns a new %s instance\nfunc New(client ethclient.Client, addr ethtypes.Address) (%s, error) {\n\tcontract, err := client.Contract(addr, []byte(ABI))\n\tif err != nil {\n\t\treturn nil, fmt.Errorf(\"failed to bind %s at address=%s\", addr, err)\n\t}\n\treturn &%s{\n\t\tcontract: contract,\n\t}, nil\n}\n\n%s\n", structName(contractName), contractABI, contractBIN, strings.Title(contractName), strings.Title(contractName), contractName, "%v: %v", structName(contractName), buildDeployer(contractName, constructorSig))
}

func buildDeployer(contractName string, constructorSig method) string {
	return fmt.Sprintf(`// Deploy a new %s contract, the contract returned by this function can only be used after signing and submitting the tx returned by this function.
func Deploy(ctx context.Context, client ethclient.Client%s) (ethtypes.Tx, %s, error) {
	contractBin, err := hex.DecodeString(BIN)
	if err != nil {
		return ethtypes.Tx{}, nil, err
	}
	tx, contract, err := ethtypes.DeployContract(ctx, client.EthClient(), []byte(ABI), contractBin, signer%s)
	if err != nil {
		return ethtypes.Tx{}, nil, err
	}
	return tx, &%s{contract}, nil
}`, strings.Title(contractName), formatInArgs(constructorSig.Constant, constructorSig.Payable, constructorSig.Inputs), strings.Title(contractName), buildParams(constructorSig), structName(contractName))
}

func buildFunctions(contractName string, contractABI contractABI) string {
	funcString := ""
	for _, method := range contractABI {
		if method.Type == "function" {
			funcString = fmt.Sprintf("%sfunc (c *%s) %s%s\n\n", funcString, structName(contractName), functionSig(method), functionBody(method))
		}
		if method.Type == "event" {
			funcString = fmt.Sprintf("%sfunc (c *%s) %s%s\n\n", funcString, structName(contractName), functionEventSig(method), functionEventBody(method))
		}
	}
	funcString = fmt.Sprintf("%sfunc (c *%s) %s%s\n", funcString, structName(contractName), "Address() ethtypes.Address",
		`{
			return c.contract.Address()
		}`)
	return funcString
}

func buildInterface(contractName string, contractABI contractABI) string {
	interfaceString := fmt.Sprintf("type %s interface{\n", strings.Title(contractName))
	for _, method := range contractABI {
		if method.Type == "function" {
			interfaceString = fmt.Sprintf("%s\t%s\n", interfaceString, functionSig(method))
		}
		if method.Type == "event" {
			interfaceString = fmt.Sprintf("%s\t%s\n", interfaceString, functionEventSig(method))
		}
	}
	interfaceString = fmt.Sprintf("%s\t%s\n", interfaceString, "Address() ethtypes.Address")
	return fmt.Sprintf("%s}\n\n", interfaceString)
}

func functionSig(method method) string {
	return fmt.Sprintf("%s(ctx context.Context%s) (%serror)", strings.Title(method.Name), formatInArgs(method.Constant, method.Payable, method.Inputs), formatOutArgs(method.Constant, method.Outputs))
}

func functionEventSig(method method) string {
	return fmt.Sprintf("Watch%s(ctx context.Context%s) (error)", strings.Title(method.Name), formatEventInArgs(method.Inputs))
}

func functionBody(method method) string {
	params := buildParams(method)
	if !method.Constant {
		return fmt.Sprintf("{\n\treturn c.contract.BuildTx(ctx, signer, \"%s\"%s)\n}", method.Name, params)
	}
	var declaration, returnValue string
	if len(method.Outputs) == 1 {
		declaration = fmt.Sprintf("\n\targ := new(%s)", method.Outputs[0].GoType)
		returnValue = fmt.Sprintf("\n\treturn *arg, nil\n")
	} else {
		if sameType(method.Outputs) {
			declaration = fmt.Sprintf("\n\targ := new([]%s)", method.Outputs[0].GoType)
			outArgs := ""
			for i := range method.Outputs {
				outArgs = fmt.Sprintf("%sret[%d], ", outArgs, i)
			}
			returnValue = fmt.Sprintf("\n\tret := *arg\n\treturn %snil\n", outArgs)
		} else {
			panic("multi-type return values not supported")
		}
	}
	return fmt.Sprintf("{%s\n\tif err := c.contract.Call(ctx, ethtypes.Address{}, arg, \"%s\"%s); err != nil{\n\t\treturn %s\n\t}%s}", declaration, method.Name, params, errVals(method.Outputs), returnValue)
}

func buildParams(method method) string {
	params := ""
	for i, in := range method.Inputs {
		if in.Name == "" {
			in.Name = fmt.Sprintf("arg%d", i)
		}
		if in.Name[:1] == "_" {
			in.Name = in.Name[1:]
		}
		params = fmt.Sprintf("%s, %s", params, in.Name)
	}
	if method.Payable {
		return fmt.Sprintf(", msgValue%s", params)
	}
	if !method.Constant {
		return fmt.Sprintf(", nil%s", params)
	}
	return params
}

func functionEventBody(method method) string {
	indexedArgs := indexedArgs(method)
	params := ""
	if len(indexedArgs) != 0 {
		params = fmt.Sprintf("\n\ttopics := [%d]ethtypes.Hash{}", len(indexedArgs))
		for k, in := range indexedArgs {
			if in.Name[:1] == "_" {
				in.Name = in.Name[1:]
			}
			params = fmt.Sprintf("%s\n\tif (%s != %s) {\n\t\ttopics[%d] = %s\n\t}", params, in.Name, defaultValue(in.GoType), k, convertToHex(in.Name, in.GoType))
		}
	}

	args := ""
	if params != "" {
		args = ", topics[:]..."
	}
	return fmt.Sprintf("{%s\n\treturn c.contract.Watch(ctx, events, beginBlockNum, \"%s\"%s)\n}", params, method.Name, args)
}

func indexedArgs(method method) []arg {
	indexedArgs := []arg{}
	for _, in := range method.Inputs {
		if !in.Indexed {
			continue
		}
		indexedArgs = append(indexedArgs, in)
	}
	return indexedArgs
}

func formatInArgs(constant, payable bool, inArgs []arg) string {
	var formattedArgs string
	for i, inArg := range inArgs {
		if inArg.Name == "" {
			inArg.Name = fmt.Sprintf("arg%d", i)
		}
		if inArg.Name[:1] == "_" {
			inArg.Name = inArg.Name[1:]
		}
		formattedArgs = fmt.Sprintf("%s, %s %s", formattedArgs, inArg.Name, inArg.GoType)
	}
	if !constant {
		formattedArgs = fmt.Sprintf("%s, signer ethtypes.Address", formattedArgs)
	}
	if payable {
		formattedArgs = fmt.Sprintf("%s, msgValue *big.Int", formattedArgs)
	}
	return formattedArgs
}

func formatEventInArgs(inArgs []arg) string {
	formattedArgs := ", beginBlockNum *big.Int, events chan<- ethtypes.Event"
	for _, inArg := range inArgs {
		if !inArg.Indexed {
			continue
		}
		if inArg.Name[:1] == "_" {
			inArg.Name = inArg.Name[1:]
		}
		formattedArgs = fmt.Sprintf("%s, %s %s", formattedArgs, inArg.Name, inArg.GoType)
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
	if typeString == "[]common.Address" {
		return "[]ethtypes.Address"
	}
	if len(typeString) >= 7 && typeString[len(typeString)-6:] == "]uint8" {
		return typeString[:len(typeString)-6] + "]byte"
	}
	return typeString
}

func errVals(args []arg) string {
	errMsg := ""
	for _, arg := range args {
		errMsg = fmt.Sprintf("%s%s, ", errMsg, defaultValue(arg.GoType))
	}
	return fmt.Sprintf("%serr", errMsg)
}

func defaultValue(goType string) string {
	if len(goType) >= 3 && goType[:3] == "int" || len(goType) >= 4 && goType[:4] == "uint" {
		return "0"
	}
	if len(goType) >= 2 && goType[:2] == "[]" {
		return "nil"
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

func convertToHex(inName, goType string) string {
	if goType[:4] == "uint" || goType[:3] == "int" {
		return fmt.Sprintf("ethtypes.HashFromHex(big.NewInt(int(%s)).Text(16))", inName)
	}
	if len(goType) >= 6 && (goType[len(goType)-5:] == "]byte") {
		return fmt.Sprintf("ethtypes.HashFromHex(fmt.Sprintf(\"%s\", %s[:]))", "%x", inName)
	}
	switch goType {
	case "string", "[]byte", "[][]byte":
		return fmt.Sprintf("ethtypes.Keccak256(%s)", inName)
	case "*big.Int":
		return fmt.Sprintf("ethtypes.HashFromHex(%s.Text(16))", inName)
	case "ethtypes.Address":
		return fmt.Sprintf("ethtypes.HashFromHex(%s.Hex())", inName)
	default:
		panic(fmt.Sprintf("unsupported go type: %s", goType))
	}
}

func sameType(outs []arg) bool {
	if len(outs) <= 1 {
		return true
	}
	outType := outs[0].Type
	for _, out := range outs {
		if out.Type != outType {
			return false
		}
	}
	return true
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
	Indexed bool   `json:"indexed,omitempty"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	GoType  string
}

type contractABI []method

func newABI(abiString string) (contractABI, method, error) {
	ethABI, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		return nil, method{}, err
	}
	cABI := contractABI{}
	constructorMethod := method{}
	json.Unmarshal([]byte(abiString), &cABI)
	for i, method := range cABI {
		if method.Type == "function" {
			for j, in := range ethABI.Methods[method.Name].Inputs {
				cABI[i].Inputs[j].GoType = formatType(in.Type)
			}
			for j, out := range ethABI.Methods[method.Name].Outputs {
				cABI[i].Outputs[j].GoType = formatType(out.Type)
			}
		}
		if method.Type == "event" {
			for j, in := range ethABI.Events[method.Name].Inputs {
				cABI[i].Inputs[j].GoType = formatType(in.Type)
			}
		}
		if method.Type == "constructor" {
			for j, in := range ethABI.Constructor.Inputs {
				cABI[i].Inputs[j].GoType = formatType(in.Type)
			}
			constructorMethod = cABI[i]
		}
	}
	return cABI, constructorMethod, nil
}

func structName(contractName string) string {
	if contractName[0] >= 65 && contractName[0] <= 90 {
		return string(append([]byte{contractName[0] + byte(0x20)}, contractName[1:]...))
	}
	return contractName
}

func getBin(address ethtypes.Address, network string) string {
	var url string
	switch network {
	case "kovan":
		url = fmt.Sprintf("https://kovan.etherscan.io/address/%s", address.Hex())
	case "mainnet":
		url = fmt.Sprintf("https://etherscan.io/address/%s", address.Hex())
	case "ropsten":
		url = fmt.Sprintf("https://ropsten.etherscan.io/address/%s", address.Hex())
	default:
		panic("unknown network")
	}
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	doc, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	re := regexp.MustCompile(`<div id='verifiedbytecode2'>([0-9a-f]*)</div>`)
	submatchall := re.FindAllStringSubmatch(string(doc), -1)
	return submatchall[0][1]
}
