package ipfsapi

import (
	"testing"
	"fmt"
)

func TestAdd(t *testing.T) {
	obj, err := Add("../main.go")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(obj)
	b, _ := Get(obj.Hash)
	fmt.Println(string(b))
}
