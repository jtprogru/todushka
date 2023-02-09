package pkg

import (
	"encoding/json"
	"fmt"
)

func AnyToJson(t any) []byte {
	b, err := json.Marshal(t)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return b
}
