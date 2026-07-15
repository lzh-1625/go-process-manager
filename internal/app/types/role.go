package types

type Role int

const (
	RoleRoot Role = iota
	RoleAdmin
	RoleUser
	RoleGuest
)
