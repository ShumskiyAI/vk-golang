package main

import (
	"errors"
	"fmt"
)

func i2s(data interface{}, out interface{}) error {
	// TODO Код писать тут
	return errors.New("not implemented yet")
}

func main() {
	input := map[string]any{"success": true}

	type ResultInfo struct {
		Success bool
	}

	output := new(ResultInfo)

	err := i2s(input, output)
	if err != nil {
		fmt.Println("ERR =", err)
	} else {
		fmt.Println("OUTPUT =", *output)
	}
}
