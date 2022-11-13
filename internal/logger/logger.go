package logger

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// var LwInfo *myLogWriter = NewLogWriter(func(str string, color color.Attribute) {
// 	Custom(color, fmt.Sprintf("[%s] %s", s.Stage, str))
// }, color.FgCyan)

// var LwWhite *myLogWriter = NewLogWriter(func(str string, color color.Attribute) {
// 	Custom(color, fmt.Sprintf("[%s] %s", s.Stage, str))
// }, color.FgWhite)

// var LwError *myLogWriter = NewLogWriter(func(str string, color color.Attribute) {
// 	Custom(color, fmt.Sprintf("[%s] %s", s.Stage, str))
// }, color.FgRed)

// var LwSuccess *myLogWriter = NewLogWriter(func(str string, color color.Attribute) {
// 	Custom(color, fmt.Sprintf("[%s] %s", s.Stage, str))
// }, color.FgGreen)

// var LwOperation *myLogWriter = NewLogWriter(func(str string, color color.Attribute) {
// 	Custom(color, fmt.Sprintf("[%s] %s", s.Stage, str))
// }, color.FgYellow)

func Free(text ...string) {
	fmt.Println(strings.Join(text[:], " "))
}
func Custom(col color.Attribute, text ...string) {
	c := color.New(col)
	c.Println(strings.Join(text[:], " "))
}

func Info(text ...string) {
	c := color.New(color.FgCyan)
	c.Println(strings.Join(text[:], " "))
}

func Operation(text ...string) {
	c := color.New(color.FgYellow)
	c.Println(strings.Join(text[:], " "))
}

func Success(text ...string) {
	c := color.New(color.FgGreen)
	c.Println(strings.Join(text[:], " "))
}

func Error(text ...string) {
	c := color.New(color.FgRed)
	c.Println(strings.Join(text[:], " "))
}

func Fatal(text error) {
	c := color.New(color.FgRed)
	c.Println(text)
}

type myLogWriter struct {
	logFunc func(string, color.Attribute)
	line    string
	color   color.Attribute
}

func (w *myLogWriter) Write(b []byte) (int, error) {
	l := len(b)
	for len(b) != 0 {
		i := bytes.Index(b, []byte{'\n'})
		if i == -1 {
			w.line += string(b)
			break
		} else {
			w.logFunc(w.line+string(b[:i]), w.color)
			b = b[i+1:]
			w.line = ""
		}
	}

	return l, nil
}

func NewLogWriter(f func(string, color.Attribute), col color.Attribute) *myLogWriter {
	return &myLogWriter{
		logFunc: f,
		color:   col,
	}
}
