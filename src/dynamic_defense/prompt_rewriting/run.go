package promptrewriting

import (
	"covalence/src/types"
	"fmt"
)

var ID = "prompt-rewriting"

func Apply(messages *[]types.Message) error {
	fmt.Printf("Hi: %v", messages)
	return nil
}
