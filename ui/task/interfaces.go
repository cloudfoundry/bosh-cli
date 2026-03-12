package task

type Reporter interface {
	TaskStarted(int)
	TaskFinished(int, string)
	TaskOutputChunk(int, []byte)
	TaskHeartbeat(id int, state string, startedAt int64)
}

type Task interface {
	ID() int
	State() string
}
