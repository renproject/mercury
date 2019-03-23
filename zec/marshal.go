package zec

// GetUTXOResponse is the response type for get /utxo/{address} request
type GetUTXOResponse []UTXO

// GetConfirmationsResponse is the response type for get /confirmations/{txHash}
// request
type GetConfirmationsResponse int64

// GetScriptResponse is the response type for get /script/{address}
type GetScriptResponse struct {
	Status bool   `json:"status"`
	Script string `json:"script,omitempty"`
	Value  int64  `json:"value,omitempty"`
}

// PostTransactionRequest is the request type for post /tx
type PostTransactionRequest struct {
	SignedTransaction string `json:"stx"`
}
