package configuro

import (
	"fmt"

	"go.uber.org/multierr"
	"gopkg.in/go-playground/validator.v9"
)

//ErrValidationErrors Error that hold multiple errors.
type ErrValidationErrors struct {
	error
}

//ErrValidationTag Error if validation failed by a tag
type ErrValidationTag struct {
	field   string
	message string
	err     error
	tag     string
}

//ErrValidationFunc Error if validation failed by Validatable Interface.
type ErrValidationFunc struct {
	field string
	err   error
}

func newErrFieldTagValidation(field validator.FieldError, message string) *ErrValidationTag {
	return &ErrValidationTag{
		field:   field.Namespace(),
		message: message,
		err:     field.(error),
		tag:     field.Tag(),
	}
}

func newErrValidate(field string, err error) *ErrValidationFunc {
	//TODO Get field name for extra context (need to update the recursive validate function to supply the path of the err)
	return &ErrValidationFunc{
		field: field,
		err:   err,
	}
}

func (e *ErrValidationTag) Error() string {
	return fmt.Sprintf(`%s: %s`, e.field, e.message)
}

func (e *ErrValidationFunc) Error() string {
	return fmt.Sprintf(`%s`, e.err)
}

//Errors Return a list of Errors held inside ErrValidationErrors.
func (e *ErrValidationErrors) Errors() []error {
	return multierr.Errors(e.error)
}

//Unwrap to support errors IS|AS
func (e *ErrValidationTag) Unwrap() error {
	return e.err
}

//Unwrap to support errors IS|AS
func (e *ErrValidationFunc) Unwrap() error {
	return e.err
}

//Unwrap to support errors IS|AS
func (e *ErrValidationErrors) Unwrap() error {
	return e.error
}
