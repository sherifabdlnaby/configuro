package configuro

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/go-multierror"
	"gopkg.in/go-playground/validator.v9"
)

//Validatable Any Type that Implements this interface will have its Validate() function called when validating config.
type Validatable interface {
	Validate() error
}

// TODO make them interfaces ?

//ErrFieldTagValidation Error if validation failed by a tag
type ErrFieldTagValidation struct {
	field   string
	message string
	err     error
	tag     string
}

//ErrValidate Error if validation failed by Validatable Interface.
type ErrValidate struct {
	field string
	err   error
}

func newErrFieldTagValidation(field validator.FieldError, message string) *ErrFieldTagValidation {
	return &ErrFieldTagValidation{
		field:   field.Namespace(),
		message: message,
		err:     field.(error),
		tag:     field.Tag(),
	}
}

func newErrValidate(field string, err error) *ErrValidate {
	return &ErrValidate{
		field: field,
		err:   err,
	}
}

func (e *ErrFieldTagValidation) Error() string {
	return fmt.Sprintf(`%s: %s`, e.field, e.message)
}

func (e *ErrValidate) Error() string {
	//TODO Get field name for extra context (need to update the recursive validate function to supply the path of the err)
	return fmt.Sprintf(`%s`, e.err)
}

//Unwrap to support errors IS|AS
func (e *ErrFieldTagValidation) Unwrap() error {
	return e.err
}

//Unwrap to support errors IS|AS
func (e *ErrValidate) Unwrap() error {
	return e.err
}

//Validate Validate Struct using Tags and Any Fields that Implements the validatable interface.
func (c *Config) Validate(configStruct interface{}) error {

	var multierr error

	if c.validateUsingTags {
		// validate tags
		err := c.validator.Struct(configStruct)

		if err != nil {
			errorLists, ok := err.(validator.ValidationErrors)
			if ok {
				for _, err := range errorLists {
					multierr = multierror.Append(multierr, newErrFieldTagValidation(err, err.Translate(c.validatorTrans)))
				}
			} else {
				multierr = multierror.Append(multierr, err)
			}

		}
	}

	err := recursiveValidate(configStruct, c.validateRecursive, c.validateStopOnFirstErr)
	if err != nil {
		multierr = multierror.Append(multierr, err)
	}

	return multierr
}

func recursiveValidate(obj interface{}, recursive bool, returnOnFirstErr bool) error {

	var multierr error

	if obj == nil {
		return nil
	}

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Recursively Validate Obj Fields / Elements
	switch val.Kind() {
	case reflect.Struct:
		if recursive {
			for i := 0; i < val.NumField(); i++ {
				field := val.Field(i)
				if field.CanInterface() {
					err := recursiveValidate(field.Interface(), recursive, returnOnFirstErr)
					if err != nil {
						multierr = multierror.Append(multierr, err)
						if returnOnFirstErr {
							return multierr
						}
					}
				}
			}
		}
	case reflect.Map:
		for _, e := range val.MapKeys() {
			v := val.MapIndex(e)
			if v.CanInterface() {
				err := recursiveValidate(v.Interface(), recursive, returnOnFirstErr)
				if err != nil {
					multierr = multierror.Append(multierr, err)
					if returnOnFirstErr {
						return multierr
					}
				}
			}
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			v := val.Index(i)
			if v.CanInterface() {
				err := recursiveValidate(v.Interface(), recursive, returnOnFirstErr)
				if err != nil {
					multierr = multierror.Append(multierr, err)
					if returnOnFirstErr {
						return multierr
					}
				}
			}
		}
	}

	// Call Validate on the value it self
	if val.CanInterface() {
		validatable, ok := val.Interface().(Validatable)
		if ok {
			err := validatable.Validate()
			if err != nil {
				multierr = multierror.Append(multierr, newErrValidate("", err))
			}
		}
	}

	return multierr
}
