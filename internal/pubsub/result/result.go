package result

type Type int

const (
	Success Type = iota
	RetryRequire
	Discard
)
