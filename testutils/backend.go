package testutils

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func CreateSimulatedEthNetwork() (*bind.TransactOpts, *backends.SimulatedBackend, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate private key: %v", err)
	}
	auth := bind.NewKeyedTransactor(key)

	alloc := make(core.GenesisAlloc)
	alloc[auth.From] = core.GenesisAccount{Balance: big.NewInt(1000000000000000000)}
	backend := backends.NewSimulatedBackend(alloc, 10000000)

	return auth, backend, nil
}

func StartGanacheServer(ctx context.Context, key *ecdsa.PrivateKey, balance *big.Int) (cmd *exec.Cmd, err error) {
	privateKeyBytes := crypto.FromECDSA(key)
	privateKeyHexString := hexutil.Encode(privateKeyBytes)
	cmdName := "ganache-cli"
	cmdArgs := []string{fmt.Sprintf("--account=%s,%s", privateKeyHexString, balance.String())}
	cmd = exec.Command(cmdName, cmdArgs...)
	cmd.Stderr = os.Stderr
	cmd.Start()

	errorChan := make(chan error, 1)
	doneChan := make(chan bool, 1)

	client, err := ethclient.Dial("http://localhost:8545")
	hexAddr := crypto.PubkeyToAddress(key.PublicKey).Hex()

	go func() {
		b, err := client.BalanceAt(ctx, common.HexToAddress(hexAddr), nil)
		for err != nil {
			time.Sleep(1 * time.Second)
			b, err = client.BalanceAt(ctx, common.HexToAddress(hexAddr), nil)
		}
		if b.Cmp(balance) == 0 {
			doneChan <- true
		} else {
			errorChan <- fmt.Errorf("balance check failed when starting ganache-cli")
		}
	}()

	select {
	case <-ctx.Done():
		return cmd, fmt.Errorf("context deadline exceeded before ganache-cli server could start")
	case err = <-errorChan:
		return cmd, err
	case <-doneChan:
		return cmd, nil
	}
}
