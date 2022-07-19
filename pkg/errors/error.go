package errors

func NewCommandError(err error, code int) *CommandError {
	return &CommandError{
		err:  err,
		Code: code,
	}
}

type CommandError struct {
	err  error
	Code int
}

func (c *CommandError) Error() string {
	return c.err.Error()
}

func (e *CommandError) Unwrap() error {
	return e.err
}
