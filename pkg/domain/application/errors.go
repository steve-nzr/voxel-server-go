package application

type NotFoundError struct {
	Message string
}

func (e NotFoundError) Error() string {
	return e.Message
}

func NewNotFoundError(what string) NotFoundError {
	return NotFoundError{
		Message: what + " not found",
	}
}
