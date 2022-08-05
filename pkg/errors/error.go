package errors

const (
	// CodeIO is used for input/output errors such as creating or closing files.
	CodeIO = iota + 1
	// CodeCommand is used for errors that arise from tea commands logic.
	CodeCommand
	// CodeThirdParty is used for errors that are returned from third party packages.
	CodeThirdParty
	// CodeSerializing is used for marshaling/unmarshaling errors.
	CodeSerializing
	// CodeDependency is used for errors that arise from invalid tea input.
	CodeDependency
)

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

func (e *CommandError) Error() string {
	return e.err.Error()
}

func (e *CommandError) Unwrap() error {
	return e.err
}
