package utils

import (
	"encoding/json"
	"fmt"
)

func AnyToJSON(t any) []byte {
	b, err := json.Marshal(t)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return b
}
