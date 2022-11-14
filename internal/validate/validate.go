package validate

import (
	"os"

	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/logger"
	"gopkg.in/validator.v2"
)

func ValidatePopulatedActionFile(a *[]config.Action) {
	if errs := validator.Validate(&a); errs != nil {
		logger.Operation("Your Configuration has Problems:")
		logger.Error(errs.Error())
		os.Exit(1)
	}
}
