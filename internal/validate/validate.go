package validate

import (
	"errors"
	"os"
	"reflect"
	"strings"

	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/engine"
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

func ValidateRepoFile(w *config.RepoFile) {
	validator.SetValidationFunc("nowhitespace", noWhiteSpace)
	if errs := validator.Validate(&w); errs != nil {
		logger.Operation("Your Repo Config file has Problems:")
		logger.Error(errs.Error())
		os.Exit(1)
	}

}

func ValidateRepo(w *config.Repo) {
	validator.SetValidationFunc("nowhitespace", noWhiteSpace)
	if errs := validator.Validate(&w); errs != nil {
		logger.Operation("Your Repo Config file has Problems:")
		logger.Error(errs.Error())
		os.Exit(1)
	}
}

func ValidateWorkflows(w *[]engine.Workflow) {
	validator.SetValidationFunc("nowhitespace", noWhiteSpace)
	if errs := validator.Validate(&w); errs != nil {
		logger.Operation("Your Workflows have Problems:")
		logger.Error(errs.Error())
		os.Exit(1)
	}
}
