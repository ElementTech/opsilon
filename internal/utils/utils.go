package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jatalocks/opsilon/internal/internaltypes"
	"github.com/manifoldco/promptui"
)

func Confirm(act internaltypes.Workflow) (bool, error) {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Run %v", act.ID),
		IsConfirm: true,
		Default:   "y",
	}
	validate := func(s string) error {
		if len(s) == 1 && strings.Contains("YyNn", s) || prompt.Default != "" && len(s) == 0 {
			return nil
		}
		return errors.New("invalid input")
	}
	prompt.Validate = validate

	_, err := prompt.Run()
	confirmed := !errors.Is(err, promptui.ErrAbort)
	if err != nil && confirmed {
		fmt.Println("ERROR: ", err)
		return false, err
	}

	return confirmed, nil
}
