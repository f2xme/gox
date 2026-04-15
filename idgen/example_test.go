package idgen_test

import (
	"fmt"

	"github.com/f2xme/gox/idgen"
)

func ExampleSnowflake() {
	id, err := idgen.Snowflake()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Snowflake ID: %d\n", id)

	info := idgen.ParseSnowflake(id)
	fmt.Printf("Node ID: %d\n", info.NodeID)
	// Output will vary
}

func ExampleUUID() {
	id := idgen.UUID()
	fmt.Printf("UUID length: %d\n", len(id.String()))
	// Output: UUID length: 36
}

func ExampleULID() {
	id := idgen.ULID()
	fmt.Printf("ULID length: %d\n", len(id.String()))
	// Output: ULID length: 26
}

func ExampleShortID() {
	id, err := idgen.ShortID()
	if err != nil {
		panic(err)
	}
	fmt.Printf("ShortID length: %d\n", len(id))
	// Output: ShortID length: 8
}

func ExampleAutoIncrement() {
	gen := idgen.NewAutoIncrement(1000)
	fmt.Println(gen.Next())
	fmt.Println(gen.Next())
	fmt.Println(gen.Next())
	// Output:
	// 1000
	// 1001
	// 1002
}

func ExampleGeneratorFunc() {
	counter := 0
	gen := idgen.GeneratorFunc(func() string {
		counter++
		return fmt.Sprintf("ORDER-%04d", counter)
	})

	fmt.Println(gen.Generate())
	fmt.Println(gen.Generate())
	// Output:
	// ORDER-0001
	// ORDER-0002
}
