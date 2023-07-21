package checker

type timeoutError struct{}

func (e *timeoutError) Error() string {
	return "I/O timeout"
}

func (e *timeoutError) Timeout() bool {
	return true
}

func (e *timeoutError) Temporary() bool {
	return true
}
