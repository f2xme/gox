package idverify_test

import (
	"context"
	"fmt"

	"github.com/f2xme/gox/idverify"
	"github.com/f2xme/gox/idverify/adapter/mock"
)

func Example() {
	v, err := mock.New()
	if err != nil {
		panic(err)
	}

	res, err := v.Verify(context.Background(), idverify.Request{
		Name:     "张三",
		IDNumber: "110101199001011234",
	})
	if err != nil {
		fmt.Println("system error:", err)
		return
	}
	if res.Matched {
		fmt.Println("matched")
	}
	// Output: matched
}
