package errors

func NewErrCommand(err error, code int) *ErrCommand {
	return &ErrCommand{
		Err:  err,
		Code: code,
	}
}

type ErrCommand struct {
	Err  error
	Code int
}

func (c *ErrCommand) Error() string {
	return c.Err.Error()
}
