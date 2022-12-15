package common

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

func ValidateStruct(validate validator.Validate, s interface{}) error {
	// returns nil or ValidationErrors ( []FieldError )
	err := validate.Struct(s)
	if err != nil {

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil value
		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
			return err
		}

		fmt.Println(err)
		return err
	}

	return nil
}

func ValidateVar(validate validator.Validate, s interface{}, tag string) error {
	err := validate.Var(s, tag)

	if err != nil {
		fmt.Println(err) // output: Key: "" Error:Field validation for "" failed on the "" tag
		return err
	}

	return nil
}
