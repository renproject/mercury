package gsnclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/renproject/mercury/sdk/client/ethclient"
	"github.com/renproject/mercury/types/ethtypes"
	"github.com/renproject/phi"
	"github.com/sirupsen/logrus"
)

type RelayerStatusResponse struct {
	RelayServerAddress string `json:"RelayServerAddress"`
	MinGasPrice        uint64 `json:"MinGasPrice"`
	Ready              bool   `json:"Ready"`
	Version            string `json:"Version"`
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

	relayerURLs   []string
	relayerURLsMu *sync.RWMutex
}

func NewGSNClient(options GSNClientOptions) gsnClient {
	return gsnClient{
		options:       options,
		relayerURLs:   make([]string, 0, options.MaxClients),
		relayerURLsMu: new(sync.RWMutex),
	}
}

// Run creates a list of potential Relays by listening to events on the RelayHub.
func (gsnclient *gsnClient) Run(ctx context.Context, recipientAddress ethtypes.Address, recipientABI []byte, relayHubABI []byte) error {
	gsnclient.options.Logger.Infof("Start running a GSN client")
	client, err := ethclient.New(gsnclient.options.Logger, gsnclient.options.Network)
	if err != nil {
		return err
	}

	gsnRecipient, err := ethtypes.NewContract(client.EthClient(), recipientAddress, recipientABI)
	if err != nil {
		return err
	}
	gsnclient.options.Logger.Infof("Obtained the GSN recipient contract on %v", gsnclient.options.Network)

	// First, the client MUST get the RelayHub used by the target gsnRecipient, by
	// calling target.getHubAddr()
	relayHubAddress := common.Address{}
	if err := gsnRecipient.Call(ctx, ethtypes.Address{}, &relayHubAddress, "getHubAddr"); err != nil {
		return err
	}
	gsnclient.options.Logger.Infof("GSN Recipient is registered with %v RelayHub", relayHubAddress.String())

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

	relayHub, err := ethtypes.NewContract(client.EthClient(), ethtypes.Address(relayHubAddress), relayHubABI)
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

					gsnclient.options.Logger.Infof("Got event %s=>%s", event.Name, event.Args["url"])
					if event.Name == RelayAdded {
						// For each relay, the client keeps the relay address and url
						gsnclient.addRelayer(event.Args["url"].(string))
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

func (gsnclient *gsnClient) addRelayer(relayerURL string) {
	gsnclient.relayerURLsMu.Lock()
	defer gsnclient.relayerURLsMu.Unlock()

	if _, ok := gsnclient.findRelayer(relayerURL); !ok && len(gsnclient.relayerURLs) < int(gsnclient.options.MaxClients) {
		gsnclient.relayerURLs = append(gsnclient.relayerURLs, relayerURL)
	}
}

func (gsnclient *gsnClient) removeRelayer(relayerURL string) {
	gsnclient.relayerURLsMu.Lock()
	defer gsnclient.relayerURLsMu.Unlock()

	if i, ok := gsnclient.findRelayer(relayerURL); ok {
		gsnclient.relayerURLs[i], gsnclient.relayerURLs = gsnclient.relayerURLs[len(gsnclient.relayerURLs)-1], gsnclient.relayerURLs[:len(gsnclient.relayerURLs)-1]
	}
}

func (gsnclient *gsnClient) findRelayer(url string) (int, bool) {
	for i, relayer := range gsnclient.relayerURLs {
		if relayer == url {
			return i, true
		}
	}
	return 0, false
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
			relayers := []string{}
			func() {
				gsnclient.relayerURLsMu.RLock()
				defer gsnclient.relayerURLsMu.RUnlock()

				relayers = gsnclient.relayerURLs
			}()

			for _, relayer := range relayers {
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

				response := RelayerStatusResponse{}
				if err := json.Unmarshal(body, &response); err != nil {
					return err
				}

				if !response.Ready {
					gsnclient.removeRelayer(relayer)
					continue
				}
				gsnclient.options.Logger.Infof("Relayer [%s]: Address = %s, MinGasPrice = %v, Ready = %v, Version: %s", relayer, response.RelayServerAddress, response.MinGasPrice, response.Ready, response.Version)
			}
		}
	}
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
// NOTE: this is a naming bug: its a senderNonce, not recpientNonce...
// RelayHubAddress: the address of the relay hub.
// relay-address: this is not sent over the protocol, but it is added to the hash to create the signature
// signature: a signature over the above parameters (see "Calculating signature" below)
// RelayMaxNonce: the maximum nonce value of the relay itself. The client should read the getTransactionCount(relayAddress), and add a "gap". a "gap" of zero means the clients will only accept the transction to be the very next tx of the relay. a larger gap means the client would accept its transction to be queued.
// Note that this parameter is NOT signed (the relay can't be penalized for putting higher nonce). But if the client sees the relay's returned transaction contains a higher nonce, it would simply re-send through another relay, and this transaction (that will getsent later) will be rejected - and the relay would pay for this rejection.
// approvalData: This is an extra data that MAY be used by custom clients and target contracts. It is not signed by the signature, and by default its empty.

// Calculating signature

// concatenate the byte values of the fields: from, to, encodedFunction, relayFee, gasPrice, gasLimit, nonce, RelayHubAddress, relayAddress
// addresses are packed as 40 bytes, uint values as 32 bytes. the encodedFunction is encoded as a byte array

// add a 4-byte prefix "rlx:"

// create a keccak256 hash of the above string

// create ethereum hash of the above:

// add a prefix "\x19Ethereum Signed Message:\n32"

// create a keccak256 hash.

// sign the generated hash with the from's field private-key

// return the 65-byte (r,s,v) signature

// 4. Send Request to the Relay.

// Before sending the request, the client MAY validate if the target would accept it, by calling the canRelay() method of the target contract.
// This way, the client doesn't have to trust the relay as to the result of the on-chain contract.
// Note that the relay itself will call the canRelay method too, before sending the request on-chain.

// In case the above canRelay call fails, it is most likely that no other relay would accept that call (e.g. the target contract doesn't accept request from this client) (see edge-case here)

// The client sends a POST request to /relay with the above JSON request, and waits for a response. The relay should creates a signed transaction, and returns it to the client.

// In case the relay doesn't answer after a reasonable network-delay time (e.g. 10 seconds), the client MAY continue and send a request to another relay. (use with care, as such situation makes the client appear to "attack" the relay)

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
