package errorx_test

import (
	"fmt"

	"github.com/f2xme/gox/errorx"
)

func ExampleNew() {
	err := errorx.New("something went wrong")
	fmt.Println(err.Error())
	// Output: something went wrong
}

func ExampleNewWithKind() {
	err := errorx.NewWithKind(errorx.KindValidation, "invalid email format")
	fmt.Println(err.Error())
	fmt.Println(err.Kind)
	// Output:
	// invalid email format
	// Validation
}

func ExampleNewCode() {
	err := errorx.NewCode("ERR001", "database connection failed")
	fmt.Println(err.Error())
	// Output: [ERR001] database connection failed
}

func ExampleWrap() {
	// Simulate an underlying error
	dbErr := fmt.Errorf("connection timeout")
	err := errorx.Wrap(dbErr, "failed to query users")
	fmt.Println(err.Error())
	// Output: failed to query users: connection timeout
}

func ExampleError_WithKind() {
	err := errorx.New("user not found").WithKind(errorx.KindNotFound)
	fmt.Println(err.Kind)
	// Output: NotFound
}

func ExampleError_WithMetadata() {
	err := errorx.New("validation failed").
		WithKind(errorx.KindValidation).
		WithMetadata("field", "email").
		WithMetadata("value", "invalid@")

	fmt.Println(err.Metadata["field"])
	// Output: email
}

func ExampleRegister() {
	// Register error codes with messages
	errorx.Register("USER_NOT_FOUND", "en", "User not found")
	errorx.Register("USER_NOT_FOUND", "zh", "用户未找到")

	// Create error with registered code
	err := errorx.NewCodeWithLang("USER_NOT_FOUND", "en")
	fmt.Println(err.Error())
	// Output: [USER_NOT_FOUND] User not found
}

func ExampleIsKind() {
	err := errorx.NewWithKind(errorx.KindValidation, "invalid input")

	if errorx.IsKind(err, errorx.KindValidation) {
		fmt.Println("This is a validation error")
	}
	// Output: This is a validation error
}

func ExampleIsRetryable() {
	timeoutErr := errorx.NewWithKind(errorx.KindTimeout, "request timeout")

	if errorx.IsRetryable(timeoutErr) {
		fmt.Println("Can retry this operation")
	}
	// Output: Can retry this operation
}

func ExampleFormat() {
	err := errorx.NewCode("VAL001", "validation failed").
		WithKind(errorx.KindValidation).
		WithMetadata("field", "email")

	formatted := errorx.Format(err)
	fmt.Println(len(formatted) > 20)
	// Output: true
}
