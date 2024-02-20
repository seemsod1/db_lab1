package error

type Error struct {
	Err string
}

const FailedToOptimize = "failed to optimize file"
const AlreadyOptimized = "already optimized"
const FailedToClose = "failed to close file"

func (e *Error) Error() string {
	return e.Err
}
