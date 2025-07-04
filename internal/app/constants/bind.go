package constants

const (
	BIND_NONE          = 0
	BIND_OPTION_HEADER = 1 << iota
	BIND_OPTION_BODY
	BIND_OPTION_QUERY
)
