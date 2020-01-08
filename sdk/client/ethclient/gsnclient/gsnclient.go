package gsnclient

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/renproject/mercury/sdk/account/ethaccount"
	"github.com/renproject/mercury/sdk/client/ethclient"
	"github.com/renproject/mercury/types/ethtypes"
	"github.com/renproject/phi"
	"github.com/sirupsen/logrus"
)

type RelayerData struct {
	RelayServerAddress string `json:"RelayServerAddress"`
	MinGasPrice        uint64 `json:"MinGasPrice"`
	Ready              bool   `json:"Ready"`
	Version            string `json:"Version"`
}

type Relayer struct {
	Info           RelayerData
	TransactionFee *big.Int
}

const (
	RelayAdded   = "RelayAdded"
	RelayRemoved = "RelayRemoved"
)

type GSNClientOptions struct {
	Logger              logrus.FieldLogger
	Network             ethtypes.Network
	MaxClients          uint8
	PingRelayerInterval time.Duration
}

type gsnClient struct {
	options GSNClientOptions

	relayerURLs   map[string]Relayer
	relayerURLsMu *sync.RWMutex

	relayHubABI      []byte
	relayHubAddress  ethtypes.Address
	recipientABI     []byte
	recipientAddress ethtypes.Address
}

func NewGSNClient(ctx context.Context, options GSNClientOptions, recipientAddress ethtypes.Address, recipientABI []byte, relayHubABI []byte) (gsnClient, error) {
	gsnclient := gsnClient{
		options: options,

		relayerURLs:   make(map[string]Relayer, options.MaxClients),
		relayerURLsMu: new(sync.RWMutex),

		relayHubABI:      relayHubABI,
		recipientABI:     recipientABI,
		recipientAddress: recipientAddress,
	}

	gsnRecipient, err := gsnclient.getGSNRecipient()
	if err != nil {
		return gsnClient{}, err
	}

	options.Logger.Infof("Obtained the GSN recipient contract on %v", options.Network)

	// The client MUST get the RelayHub used by the target gsnRecipient, by
	// calling target.getHubAddr()
	relayHubAddress := common.Address{}
	if err := gsnRecipient.Call(ctx, ethtypes.Address{}, &relayHubAddress, "getHubAddr"); err != nil {
		return gsnClient{}, err
	}
	gsnclient.relayHubAddress = ethtypes.Address(relayHubAddress)

	return gsnclient, nil
}

// Run creates a list of potential Relays by listening to events on the RelayHub.
func (gsnclient *gsnClient) Run(ctx context.Context) error {
	gsnclient.options.Logger.Infof("Start running a GSN client")
	client, err := ethclient.New(gsnclient.options.Logger, gsnclient.options.Network)
	if err != nil {
		return err
	}

	// Next, the client SHOULD filter for recent RelayAdded and RelayRemoved
	// events sent by the RelayHub. Since a relay is required to send such event
	// every 6000 blocks (roughly 24hours), the client should look at most at
	// the latest 6000 blocks. For each relay, look for the latest RelayAdded
	// event.
	currentBlockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		return err
	}
	gsnclient.options.Logger.Infof("Current block number: %v", currentBlockNumber)

	relayHub, err := ethtypes.NewContract(client.EthClient(), gsnclient.relayHubAddress, gsnclient.relayHubABI)
	if err != nil {
		return err
	}

	events := make(chan ethtypes.Event)

	gsnclient.options.Logger.Infof("Start listening to RelayAdded and RelayRemoved events on the RelayHub")

	phi.ParBegin(
		func() {
			if err := relayHub.Watch(ctx, events, currentBlockNumber.Sub(currentBlockNumber, big.NewInt(6000)), RelayAdded); err != nil {
				gsnclient.options.Logger.Errorf("failed to watch for relay added events: %v", err)
				return
			}
		},
		func() {
			if err := relayHub.Watch(ctx, events, currentBlockNumber.Sub(currentBlockNumber, big.NewInt(6000)), RelayRemoved); err != nil {
				gsnclient.options.Logger.Errorf("failed to watch for relay removed events: %v", err)
				return
			}
		},
		func() {
			defer close(events)

			for {
				select {
				case <-ctx.Done():
					return
				case event, ok := <-events:
					if !ok {
						return
					}

					gsnclient.options.Logger.Infof("Got event %s=>[url:%s, txFees:%v]\n", event.Name, event.Args["url"], event.Args["transactionFee"].(*big.Int))
					if event.Name == RelayAdded {
						// For each relay, the client keeps the relay address and url
						gsnclient.addRelayer(event.Args["url"].(string), event.Args["transactionFee"].(*big.Int))
					} else {
						// Relays with RelayRemoved event should be removed from the list.
						gsnclient.removeRelayer(event.Args["url"].(string))
					}

					// TODO: Now the client should filter out relays that it doesn't care about.
					// e.g.: Ignore relays with stake time or stake delay below a given
					// threshold. Sort the relays in the preferred order. e.g.: prefer relays
					// with lower transaction fee, and also depend on the Relay Reputation
					// Dynamic relay selection: before making the call, the client SHOULD "ping"
					// the relay (see below)
				}
			}
		},
	)
	return nil
}

func (gsnclient *gsnClient) getGSNRecipient() (ethtypes.Contract, error) {
	client, err := ethclient.New(gsnclient.options.Logger, gsnclient.options.Network)
	if err != nil {
		return nil, err
	}

	return ethtypes.NewContract(client.EthClient(), gsnclient.recipientAddress, gsnclient.recipientABI)
}

func (gsnclient *gsnClient) addRelayer(relayerURL string, txFees *big.Int) {
	gsnclient.relayerURLsMu.Lock()
	defer gsnclient.relayerURLsMu.Unlock()

	if _, ok := gsnclient.relayerURLs[relayerURL]; !ok && len(gsnclient.relayerURLs) < int(gsnclient.options.MaxClients) {
		gsnclient.relayerURLs[relayerURL] = Relayer{TransactionFee: txFees}
	}
}

func (gsnclient *gsnClient) updateRelayer(relayerURL string, relayerData RelayerData) {
	gsnclient.relayerURLsMu.Lock()
	defer gsnclient.relayerURLsMu.Unlock()

	relayer, ok := gsnclient.relayerURLs[relayerURL]
	if !ok {
		return
	}
	relayer.Info = relayerData
	gsnclient.relayerURLs[relayerURL] = relayer
}

func (gsnclient *gsnClient) removeRelayer(relayerURL string) {
	gsnclient.relayerURLsMu.Lock()
	defer gsnclient.relayerURLsMu.Unlock()

	delete(gsnclient.relayerURLs, relayerURL)
}

// 2. Select a Relay For each potential relay, the client "pings" the relay by
// sending a /getaddr request. Validate the relay is valid (contains Ready:
// true) Validate the relay supports this protocol: version:0.4.x Validate the
// MinGasPrice: The relay MAY reject request with lower gas-price, so the client
// SHOULD skip requesting the relay if the relay requires higher gas-price.
func (gsnclient *gsnClient) PingRelayers(ctx context.Context) error {
	ticker := time.NewTicker(gsnclient.options.PingRelayerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			relayers := map[string]Relayer{}
			func() {
				gsnclient.relayerURLsMu.RLock()
				defer gsnclient.relayerURLsMu.RUnlock()

				relayers = gsnclient.relayerURLs
			}()

			for relayer := range relayers {
				body, err := func() ([]byte, error) {
					resp, err := http.Get(fmt.Sprintf("%s/getaddr", relayer))
					if err != nil {
						return nil, err
					}

					defer resp.Body.Close()

					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						return nil, err
					}
					return body, nil
				}()
				if body == nil || err != nil {
					return err
				}

				fmt.Printf("%s\n", body)
				response := RelayerData{}
				if err := json.Unmarshal(body, &response); err != nil {
					return err
				}

				if !response.Ready {
					gsnclient.removeRelayer(relayer)
					continue
				}
				gsnclient.updateRelayer(relayer, response)
				gsnclient.options.Logger.Infof("Relayer [%s]: Address = %s, MinGasPrice = %v, Ready = %v, Version: %s", relayer, response.RelayServerAddress, response.MinGasPrice, response.Ready, response.Version)
			}
		}
	}
}

type GSNTx struct {
	From            ethtypes.Address
	To              ethtypes.Address
	EncodedFn       []byte
	GasPrice        *big.Int
	GasLimit        *big.Int
	RecipientNonce  *big.Int
	RelayHubAddress ethtypes.Address
	RelayAddress    ethtypes.Address
	Signature       []byte
	RelayMaxNonce   *big.Int
	ApprovalData    []byte
	RelayFee        *big.Int
}

type GSNTxReq struct {
	From            string `json:"from"`
	To              string `json:"to"`
	EncodedFn       string `json:"encodedFunction"`
	GasPrice        uint64 `json:"gasPrice"`
	GasLimit        uint64 `json:"gasLimit"`
	RecipientNonce  uint64 `json:"RecipientNonce"`
	RelayHubAddress string `json:"RelayHubAddress"`
	RelayAddress    string `json:"relay-address"`
	Signature       string `json:"signature"`
	RelayMaxNonce   uint64 `json:"RelayMaxNonce"`
	ApprovalData    string `json:"approvalData"`
	RelayFee        uint64 `json:"relayFee`
}

// 3. Create a Request

// The client should create and sign a relay request, which MUST contain the following fields:
// from: the client's address
// to: the target contract's address
// encodedFunction - the function to call on the target contract (selector and params)
// relayFee: the fee the client would pay for the relay. The fee is precent above the real transaction price, so "70" means the actual fee the client will pay would be: usedGas*(100+relayFee)/100
// gasPrice: the Minimum gas price for the request. The relay MAY use higher gas-price, but will only get compensated for this advertised gas-price.
// gasLimit: the Minimum gas-limit available for the encodedFunction. Note that the actual request will have higher gas-limit, to compensate for the pre- and post- relay calls, but these are limited by a total of 250,000 gas.
// RecipientNonce: the client should put relayHub.getNonce(from) in this field.
// NOTE: this is a naming bug: its a senderNonce, not recipientNonce...
// RelayHubAddress: the address of the relay hub.
// relay-address: this is not sent over the protocol, but it is added to the hash to create the signature
// signature: a signature over the above parameters (see "Calculating signature" below)
// RelayMaxNonce: the maximum nonce value of the relay itself. The client should read the getTransactionCount(relayAddress), and add a "gap". a "gap" of zero means the clients will only accept the transction to be the very next tx of the relay. a larger gap means the client would accept its transction to be queued.
// Note that this parameter is NOT signed (the relay can't be penalized for putting higher nonce). But if the client sees the relay's returned transaction contains a higher nonce, it would simply re-send through another relay, and this transaction (that will getsent later) will be rejected - and the relay would pay for this rejection.
// approvalData: This is an extra data that MAY be used by custom clients and target contracts. It is not signed by the signature, and by default its empty.
func (gsnclient *gsnClient) PublishTx(ctx context.Context, tx ethtypes.Tx, account ethaccount.Account) (ethtypes.TxHash, error) {
	_, relayer, ok := gsnclient.getRelayer()
	if !ok {
		return ethtypes.TxHash{}, fmt.Errorf("no relayers found")
	}
	gsnTx := GSNTx{
		From:            account.Address(),
		To:              gsnclient.recipientAddress,
		EncodedFn:       tx.ToTransaction().Data(),
		RelayFee:        relayer.TransactionFee,
		GasPrice:        big.NewInt(0).SetUint64(relayer.Info.MinGasPrice),
		GasLimit:        big.NewInt(0).SetUint64(tx.ToTransaction().Gas()),
		RelayAddress:    ethtypes.AddressFromHex(relayer.Info.RelayServerAddress),
		RelayHubAddress: gsnclient.relayHubAddress,
	}
	var err error
	gsnTx.RelayMaxNonce, err = gsnclient.GetRelayNonce(ctx, ethtypes.AddressFromHex(relayer.Info.RelayServerAddress))
	if err != nil {
		return ethtypes.TxHash{}, err
	}
	gsnTx.RecipientNonce, err = gsnclient.getRelayHubNonce(ctx, account.Address())
	if err != nil {
		return ethtypes.TxHash{}, err
	}
	gsnTx.Signature, err = gsnclient.calculateSig(gsnTx, account.PrivateKey())
	if err != nil {
		return ethtypes.TxHash{}, err
	}

	gsnclient.options.Logger.Infof("From: %s\nTo: %s\nEncodedFn: %v\nGasPrice: %v\nGasLimit: %v\nRecipientNonce: %v\nRelayHubAddress: %s\nRelayAddress: %s\nSignature: %v\nRelayMaxNonce: %v\nApprovalData: %s\nRelayFee: %v\n", gsnTx.From.Hex(), gsnTx.To.Hex(), base64.StdEncoding.EncodeToString(gsnTx.EncodedFn), gsnTx.GasPrice, gsnTx.GasLimit, gsnTx.RecipientNonce, gsnTx.RelayHubAddress.Hex(), gsnTx.RelayAddress.Hex(), base64.StdEncoding.EncodeToString(gsnTx.Signature), gsnTx.RelayMaxNonce, gsnTx.ApprovalData, gsnTx.RelayFee)

	// gsnclient.sendTxToRelayer(ctx, GSNTxReq{}, relayerURL)
	return ethtypes.TxHash{}, nil
}

// Calculating signature

// concatenate the byte values of the fields: from, to, encodedFunction,
// relayFee, gasPrice, gasLimit, nonce, RelayHubAddress, relayAddress addresses
// are packed as 40 bytes, uint values as 32 bytes. the encodedFunction is
// encoded as a byte array add a 4-byte prefix "rlx:" create a keccak256 hash of
// the above string create ethereum hash of the above: add a prefix
// "\x19Ethereum Signed Message:\n32" create a keccak256 hash. sign the
// generated hash with the from's field private-key return the 65-byte (r,s,v)
// signature
func (gsnclient *gsnClient) calculateSig(tx GSNTx, privKey *ecdsa.PrivateKey) ([]byte, error) {
	// txDataStr := bytesToHexWithNoPrefix(tx.EncodedFn)
	// dataToHash := hex.EncodeToString([]byte("rlx:")) +
	// 	tx.From.Hex()[2:] +
	// 	tx.To.Hex()[2:] +
	// 	bytesToHexWithNoPrefix(tx.EncodedFn) +
	// 	uintToHexWithNoPrefix(tx.RelayFee.Int64(), 32) +
	// 	uintToHexWithNoPrefix(tx.GasPrice.Int64(), 32) +
	// 	uintToHexWithNoPrefix(tx.GasLimit.Int64(), 32) +
	// 	uintToHexWithNoPrefix(tx.RecipientNonce.Int64(), 32) +
	// 	tx.RelayHubAddress.Hex()[2:] +
	// 	tx.RelayAddress.Hex()[2:]
	dataToHash := "0x" + hex.EncodeToString([]byte("rlx:")) +
		tx.From.Hex()[2:] +
		tx.To.Hex()[2:] +
		bytesToHexWithNoPrefix(tx.EncodedFn) +
		common.BigToHash(tx.RelayFee).Hex()[2:] +
		common.BigToHash(tx.GasPrice).Hex()[2:] +
		common.BigToHash(tx.GasLimit).Hex()[2:] +
		common.BigToHash(tx.RecipientNonce).Hex()[2:] +
		tx.RelayHubAddress.Hex()[2:] +
		tx.RelayAddress.Hex()[2:]

	// fmt.Println("data", dataToHash)
	// appended := []byte{}
	// appended = append(appended, []byte("rlx:")...)

	// // TODO:
	// // 1. Try converting addresses to hex and then to ascii to get 40 bytes
	// // 2. Use as 20 byte arrays
	// // 3. Convert to hex and use that as the 40 byte array
	// // fromB := [40]byte{}
	// // copy(fromB[:], common.FromHex(tx.From.Hex()))
	// // appended = append(appended, fromB[:]...)
	// fmt.Println(string(tx.From.Hex()[2:]), tx.From.Hex(), len([]byte(tx.From.Hex()[2:])))
	// appended = append(appended, []byte(tx.From.Hex()[2:])[:]...)

	// // toB := [40]byte{}
	// // copy(toB[:], common.FromHex(tx.To.Hex()))
	// // appended = append(appended, toB[:]...)
	// appended = append(appended, []byte(tx.To.Hex()[2:])[:]...)

	// appended = append(appended, bytesToHexWithNoPrefix(tx.EncodedFn)...)

	// relayFee := [32]byte{}
	// copy(relayFee[:], tx.RelayFee.Bytes())
	// appended = append(appended, relayFee[:]...)

	// gasPrice := [32]byte{}
	// copy(gasPrice[:], tx.GasPrice.Bytes())
	// appended = append(appended, gasPrice[:]...)

	// gasLimit := [32]byte{}
	// copy(gasLimit[:], tx.GasLimit.Bytes())
	// appended = append(appended, gasLimit[:]...)

	// nonce := [32]byte{}
	// copy(nonce[:], tx.RecipientNonce.Bytes())
	// appended = append(appended, nonce[:]...)

	// // relayHubAddressB := [40]byte{}
	// // copy(relayHubAddressB[:], common.FromHex(tx.RelayHubAddress.Hex()))
	// // appended = append(appended, relayHubAddressB[:]...)
	// appended = append(appended, []byte(tx.RelayHubAddress.Hex()[2:])[:]...)

	// // relayAddressB := [40]byte{}
	// // copy(relayAddressB[:], common.FromHex(tx.RelayAddress.Hex()))
	// // appended = append(appended, relayAddressB[:]...)
	// appended = append(appended, []byte(tx.RelayAddress.Hex()[2:])[:]...)

	hash := ethtypes.Keccak256(dataToHash)
	// fmt.Println(hex.EncodeToString(hash[:]))
	// appended = []byte("\x19Ethereum Signed Message:\n32")
	// appended = append(appended, common.Hash(hash).Bytes()...)
	dataToHash = "0x" + hex.EncodeToString([]byte("\x19Ethereum Signed Message:\n32")) + hex.EncodeToString(hash[:])

	hash = ethtypes.Keccak256(dataToHash)

	return crypto.Sign(hash[:], privKey)
}

type canRelayResponse struct {
	status       uint64
	recipientCtx []byte
}

// 4. Send Request to the Relay.

// Before sending the request, the client MAY validate if the target would
// accept it, by calling the canRelay() method of the target contract. This way,
// the client doesn't have to trust the relay as to the result of the on-chain
// contract. Note that the relay itself will call the canRelay method too,
// before sending the request on-chain.

// In case the above canRelay call fails, it is most likely that no other relay
// would accept that call (e.g. the target contract doesn't accept request from
// this client) (see edge-case here)

// The client sends a POST request to /relay with the above JSON request, and
// waits for a response. The relay should creates a signed transaction, and
// returns it to the client.

// In case the relay doesn't answer after a reasonable network-delay time (e.g.
// 10 seconds), the client MAY continue and send a request to another relay.
// (use with care, as such situation makes the client appear to "attack" the
// relay)
func (gsnclient *gsnClient) sendTxToRelayer(ctx context.Context, gsnTx GSNTxReq, url string) (ethtypes.TxHash, error) {
	// TODO: Fix this
	// relayHub, err := gsnclient.getRelayHub()
	// if err != nil {
	// 	return ethtypes.TxHash{}, err
	// }
	// response := canRelayResponse{}
	// relayHub.Call(ctx, ethtypes.Address{}, &response, "canRelay", gsnTx.RelayAddress, gsnTx.From, gsnTx.To, gsnTx.EncodedFn, gsnTx.RelayFee, gsnTx.GasPrice, gsnTx.GasLimit, gsnTx.RecipientNonce, gsnTx.Signature, gsnTx.ApprovalData)

	// if response.status != 0 {
	// 	return ethtypes.TxHash{}, fmt.Errorf("canRelay call got rejected with status=%v", response.status)
	// }

	resp, err := SendRequest(gsnTx, url)
	if err != nil {
		return ethtypes.TxHash{}, err
	}

	// TODO: process resp and return txHash
	panic(resp)
}

// SendRequest sends the JSON-2.0 request to the target url and returns the response and any error.
func SendRequest(request GSNTxReq, url string) (*http.Response, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	resp, err := SendRawPost(data, url)
	if err != nil {
		fmt.Printf("Sending %s to %s resulted in an error: %v\n", string(data), url, err)
		return nil, err
	}
	return resp, nil
}

// SendRawPost sends a raw bytes as a POST request to the URL specified
func SendRawPost(data []byte, url string) (*http.Response, error) {
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}
	client := newClient(10 * time.Second)
	buff := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", url, buff)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

func newClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   2 * time.Second,
				KeepAlive: 10 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 4 * time.Second,
			ResponseHeaderTimeout: 3 * time.Second,
		},
		Timeout: timeout,
	}
}

func (gsnclient *gsnClient) getRelayer() (string, Relayer, bool) {
	gsnclient.relayerURLsMu.RLock()
	defer gsnclient.relayerURLsMu.RUnlock()

	for url, relayer := range gsnclient.relayerURLs {
		if relayer.Info != (RelayerData{}) {
			return url, relayer, true
		}
	}
	return "", Relayer{}, false
}

func (gsnclient *gsnClient) getRelayHub() (ethtypes.Contract, error) {
	client, err := ethclient.New(gsnclient.options.Logger, gsnclient.options.Network)
	if err != nil {
		return nil, err
	}

	return ethtypes.NewContract(client.EthClient(), gsnclient.relayHubAddress, gsnclient.relayHubABI)
}

func (gsnclient *gsnClient) GetRelayNonce(ctx context.Context, relayAddress ethtypes.Address) (*big.Int, error) {
	client, err := ethclient.New(gsnclient.options.Logger, gsnclient.options.Network)
	if err != nil {
		return nil, err
	}

	relayNonce, err := client.PendingNonceAt(ctx, relayAddress)
	if err != nil {
		return nil, err
	}

	return big.NewInt(0).SetUint64(relayNonce), nil
}

func (gsnclient *gsnClient) getRelayHubNonce(ctx context.Context, from ethtypes.Address) (*big.Int, error) {
	relayHub, err := gsnclient.getRelayHub()
	if err != nil {
		return nil, err
	}

	var relayHubNonce uint64
	relayHub.Call(ctx, ethtypes.Address{}, &relayHubNonce, "getNonce", from)
	return big.NewInt(0).SetUint64(relayHubNonce), nil
}

// 5. Validate Relay Response

// A response from the relay is transaction JSON in the format:

// { none: '0x1',
//   gasPrice: '0x59682f000',
//   gas: '0x24a70c',
//   to: '0x123456789abcdef0123456789abcdef012345678',
//   value: '0x0',
//   input: '0x.....',
//   v: '0x1b',
//   r: '0x779c6b594da215d65b2fe2325fa9e6f1c7d801c5162c92132e9249ae6676520b',
//   s: '0x18829cd1c02f47d9981fa4c777cf647bcae1bfa84475276e6ec7e683451ae264',
//   hash: '0xc5b4fd72a73aa050ec7112e31a90477e5375611ee12665753fffeffc62166a95'
// }
// or an error:

// { error: "..." }
// When the response is received, the client should validate it, to make sure its transaction was properly sent to the RelayHub:

// Decode the transaction to make sure its a valid ethereum transaction, signed by the relay.

// Check that the relay has enough balance to pay for it.

// The relay's nonce on the transaction meets the expectation (that is, its not too far from current relay's nonce)

// The client MAY put the transaction on-chain. In that case, it should ignore "repeated transaction" error (since the relay itself also should put it on-chain)

// The client MAY send the request to a randomly chosen other relay through the /validate URL, instead of performing the validation steps below.

// The client should wait for the relay's nonce to get incremented to the transaction nonce.

// Then it should validate that the on-chain transaction with that nonce is indeed the transaction returned to the client (note that it may have different (higher) gas-price, but otherwise should be the same)

// If the transaction with that nonce is different, it MAY call RelayHub.penalizeRepeatedNonce() to slash the relay's stake (and gain half of it)

// 6. Handle Relay Error Responses.

// In case of error, the relay MAY return an error in the format { "error": "<message>" }

// In any such case, the client should continue to send the transaction to the next available relay.

// 7. Process Trasnaction Receipt

// Wait until the trasnaction is mined.

// Since the relay MAY mine a transaction with different transaction fee, the client should NOT wait by the transaction-hash, but instead wait either for the relay's nonce to match the returned transaction nonce, or for a TransactionRelayed/CanRelayFailed event for this relay/sender/receipient.

// Once the transaction is mined, the client SHOULD check the resulting event:

// If no event was triggered, then the transaction was reverted (on the relay's expense...). The client MAY need to re-send its request, probably through another relay.

// CanRelayFailed: This should be reported as "reverted" to the calling client. However, the revert reason is not the executed method, but rather the target contract failed to accept the request on-chain - even though the same canRelay() returned a valid response when called as a view function.

// TransactionRelayed: check the status member. if its OK, then the transaction had succeeded. otherwise, the original transaction had reverted.

// RelayedCallFailed: this value indicates that the relayed transaction was reverted.

// PreRelayedFailed: the target's preRelayedCall had reverted. the relayed function was not triggered at all.

// PostRelayedFailed: the target's postRelayedCall had reverted. as a result, the relayed function was also reverted.

// RecipientBalanceChanged: the target's balance was changed. As this might cause relay fee problems, the transaction was reverted.

func bytesToHexWithNoPrefix(b []byte) string {
	hex := hex.EncodeToString(b)
	if len(hex)%2 != 0 {
		hex = "0" + hex
	}
	return hex
}

func uintToHexWithNoPrefix(val int64, length int) string {
	valStr := strconv.FormatInt(val, 16)
	if len(valStr)%2 != 0 {
		valStr = "0" + valStr
	}
	decoded, err := hex.DecodeString(valStr)
	if err != nil {
		panic(err)
	}
	for len(decoded) < length {
		decoded = append(decoded, byte(0))
	}
	return hex.EncodeToString(decoded)
}
