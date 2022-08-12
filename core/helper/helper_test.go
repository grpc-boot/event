package helper

import (
	"fmt"
	"testing"
)

func TestIdWithIp(t *testing.T) {
	id := Id4Guest(2521877123)
	fmt.Printf("%s %d", id, len(id))
}
