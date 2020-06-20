package configuro

import (
	"reflect"

	"go.uber.org/multierr"
	"gopkg.in/go-playground/validator.v9"
)

//Validatable Any Type that Implements this interface will have its WithValidateByTags() function called when validating config.
type Validatable interface {
	Validate() error
}

//Validate Validates Struct using Tags and Any Fields that Implements the Validatable interface.
func (c *Config) Validate(configStruct interface{}) error {

	var errs error

	if c.validateUsingTags {
		errs = multierr.Append(c.validateTags(configStruct), errs)
	}

	if c.validateUsingFunc {
		err := recursiveValidate(configStruct, c.validateRecursive, c.validateFuncStopOnFirstErr)
		if err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	// cast errs to ErrValidationErrs if it is multierr (So we use the package error instead of 3rd party error type)
	if len(multierr.Errors(errs)) > 1 {
		return ErrValidationErrors{error: errs}
	}
	return errs
}

func (c *Config) validateTags(configStruct interface{}) error {
	var errs error
	// validate tags
	err := c.validator.Struct(configStruct)
	if err != nil {
		errorLists, ok := err.(validator.ValidationErrors)
		if ok {
			for _, tagErr := range errorLists {
				// Add translated Tag error.
				errs = multierr.Append(errs, newErrFieldTagValidation(tagErr, tagErr.Translate(c.validatorTrans)))
			}
		} else {
			errs = multierr.Append(errs, err)
		}
	}
	return errs
}

func recursiveValidate(obj interface{}, recursive bool, returnOnFirstErr bool) error {

	var errs error

	if reflect.ValueOf(obj).Kind() == reflect.Ptr && reflect.ValueOf(obj).IsNil() {
		return nil
	}

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Recursively WithValidateByTags Obj Fields / Elements
	switch val.Kind() {
	case reflect.Struct:
		if recursive {
			for i := 0; i < val.NumField(); i++ {
				field := val.Field(i)
				if field.CanInterface() {
					err := recursiveValidate(field.Interface(), recursive, returnOnFirstErr)
					if err != nil {
						errs = multierr.Append(errs, err)
						if returnOnFirstErr {
							return errs
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
					errs = multierr.Append(errs, err)
					if returnOnFirstErr {
						return err
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
					errs = multierr.Append(errs, err)
					if returnOnFirstErr {
						return err
					}
				}
			}
		}
	}

	// Call WithValidateByTags on the value it self
	if val.CanInterface() {
		validatable, ok := val.Interface().(Validatable)
		if ok {
			err := validatable.Validate()
			if err != nil {
				errs = multierr.Append(errs, newErrValidate("", err))
			}
		}
	}

	return errs
}
