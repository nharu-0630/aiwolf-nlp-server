package model

type Status string

const (
	S_ALIVE Status = "ALIVE" // 生存
	S_DEAD  Status = "DEAD"  // 死亡
)

func (s Status) String() string {
	return string(s)
}
