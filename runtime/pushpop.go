package runtime

type PushPopType int

const (
	PushPopTunnel PushPopType = iota
	PushPopFunction
	PushPopFunctionEvaluationFromGame
)
