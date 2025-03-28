package util

import (
	"encoding/json"
	"fmt"
)

func PrettyPrint(data interface{}) (output string, err error) {
	p, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return
	}
	return fmt.Sprintf("%s ", p), nil
}
