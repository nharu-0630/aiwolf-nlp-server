package model

type Status string

const (
	S_ALIVE Status = "ALIVE"
	S_DEAD  Status = "DEAD"
)

func (s Status) String() string {
	return string(s)
}
