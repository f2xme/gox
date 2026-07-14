package geo_test

import (
	"context"
	"fmt"

	"github.com/f2xme/gox/geo"
	"github.com/f2xme/gox/geo/adapter/memory"
)

func ExampleLocator() {
	locator, err := memory.New(
		memory.WithLocation("8.8.8.8", &geo.Location{
			Country:     "美国",
			CountryCode: "US",
			Province:    "加利福尼亚",
			City:        "山景城",
			ISP:         "Google",
		}),
	)
	if err != nil {
		fmt.Println("create failed:", err)
		return
	}

	loc, err := locator.Lookup(context.Background(), "8.8.8.8")
	if err != nil {
		fmt.Println("lookup failed:", err)
		return
	}
	fmt.Println(loc.CountryCode, loc.City)
	// Output:
	// US 山景城
}
