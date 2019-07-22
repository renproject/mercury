package types

type AccessLevel uint8

const (
	FullAccess   AccessLevel = 2
	CachedAccess AccessLevel = 1
	NoAccess     AccessLevel = 0
)
