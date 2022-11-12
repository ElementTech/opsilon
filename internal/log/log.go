package log

import (
	"strings"

	"github.com/fatih/color"
)

func Info(text ...string) {
	c := color.New(color.FgCyan)
	c.Println(strings.Join(text[:], " "))
}
