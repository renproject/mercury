package bch

type Codec struct {
	alphabet      string
	reverseLookup map[rune]byte
}

func NewCodec(alphabet string) Codec {
	reverseLookup := map[rune]byte{}
	for i, char := range alphabet {
		reverseLookup[char] = byte(i)
	}
	return Codec{alphabet, reverseLookup}
}

func (codec Codec) EncodeToString(data []byte) string {
	dataStr := ""
	for _, ele := range data {
		dataStr += string(codec.alphabet[ele])
	}
	return dataStr
}

func (codec Codec) DecodeString(data string) []byte {
	d := []byte{}
	for _, b := range data {
		d = append(d, codec.reverseLookup[b])
	}
	return d
}
