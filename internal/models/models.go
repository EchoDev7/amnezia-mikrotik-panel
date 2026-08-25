package models

import "time"

type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusExpired  UserStatus = "expired"
	StatusLimited  UserStatus = "limited"
	StatusDisabled UserStatus = "disabled"
)

type User struct {
	ID               string
	Name             string
	PublicKey        string
	PresharedKey     string
	AllowedIPs       string
	Status           UserStatus
	DataLimitBytes   int64
	TotalBytes       int64
	SessionStartRxTx int64
	ExpiresAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time

	// AmneziaWG 3.1 Obfuscation Parameters
	Jc   int
	Jmin int
	Jmax int
	S1   int
	S2   int
	S3   int
	S4   int
	H1   uint32
	H2   uint32
	H3   uint32
	H4   uint32
}
