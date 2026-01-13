package helper

import (
	"encoding/json"
	"fmt"
)

func PrintStructJson(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("❌ marshal error:", err)
		return
	}
	fmt.Println(string(b))
}
