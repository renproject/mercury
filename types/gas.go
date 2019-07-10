package types

// TxSpeed indicates the tier of speed that the transaction falls under while writing to the blockchain.
type TxSpeed uint8

// TxSpeed values.
const (
	Nil = TxSpeed(iota)
	Slow
	Standard
	Fast
)
