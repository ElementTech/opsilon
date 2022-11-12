package convert

import (
	"encoding/json"
	"fmt"
	"strconv"
)

var boolValues = []string{
	"1",
	"t",
	"T",
	"TRUE",
	"true",
	"True",
	"0",
	"f",
	"F",
	"FALSE",
	"false",
	"False",
}

// nolint: goerr113
var errConversionError = func(v interface{}) error {
	return fmt.Errorf("cannot convert value %v (type %T)", v, v)
}

// ToInteger converts given value to integer.
func ToInteger(v interface{}) (int, error) {
	i, err := strconv.Atoi(fmt.Sprintf("%v", v))
	if err != nil {
		return i, errConversionError(v)
	}

	return i, nil
}

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}
