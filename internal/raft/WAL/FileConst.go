package wal

const (
	CommandBsize   = 3
	TimestampBsize = 8
	MaxWlogSize    = 4096
)

const (
	TypeJSON byte = iota + 1
	TypeString
	TypeInt
)
