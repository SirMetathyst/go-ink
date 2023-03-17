package runtime

type PushPopType int

const (
	Tunnel PushPopType = iota
	Function
	FunctionEvaluationFromGame
)
