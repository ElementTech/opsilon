package validate

import (
	"errors"
	"os"
	"reflect"
	"strings"

	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/logger"
	"gopkg.in/validator.v2"
)

func noWhiteSpace(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return errors.New("noWhiteSpace only validates strings")
	}
	if strings.Contains(st.String(), " ") {
		return errors.New("value cannot contain spaces")
	}
	return nil
}

func ValidatePopulatedActionFile(a *[]config.Action) {
	validator.SetValidationFunc("nowhitespace", noWhiteSpace)
	if errs := validator.Validate(&a); errs != nil {
		logger.Operation("Your Configuration has Problems:")
		logger.Error(errs.Error())
		os.Exit(1)
	}
}
