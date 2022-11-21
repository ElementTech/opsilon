package validate

import (
	"errors"
	"reflect"
	"strings"

	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/internaltypes"
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

func ValidateRepoFile(w *config.RepoFile) error {
	validator.SetValidationFunc("nowhitespace", noWhiteSpace)
	if errs := validator.Validate(&w); errs != nil {
		logger.Operation("Your Repo Config file has Problems:")
		return errs
	}
	return nil
}

func ValidateRepo(w *config.Repo) error {
	validator.SetValidationFunc("nowhitespace", noWhiteSpace)
	if errs := validator.Validate(&w); errs != nil {
		logger.Operation("Your Repo Config file has Problems:")
		return errs
	}
	return nil
}

func ValidateWorkflows(w *[]internaltypes.Workflow) error {
	validator.SetValidationFunc("nowhitespace", noWhiteSpace)
	if errs := validator.Validate(&w); errs != nil {
		logger.Operation("Your Workflows have Problems:")
		return errs
	}
	return nil
}
