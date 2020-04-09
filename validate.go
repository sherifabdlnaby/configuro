package configuro

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/go-multierror"
	"gopkg.in/go-playground/validator.v9"
)

type Validatable interface {
	Validate() error
}

func (c *Config) Validate(configStruct interface{}) error {

	var multierr error

	if c.validateUsingTags {
		// validate tags
		err := c.validator.Struct(configStruct)

		if err != nil {
			errorLists, ok := err.(validator.ValidationErrors)
			if ok {
				errs := errorLists.Translate(c.validatorTrans)
				for _, value := range errs {
					multierr = multierror.Append(multierr, fmt.Errorf(value))
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
				multierr = multierror.Append(multierr, err)
			}
		}
	}

	return multierr
}
