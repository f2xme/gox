package email_test

import (
	"fmt"

	"github.com/f2xme/gox/email"
)

func ExampleNew() {
	client, err := email.New("smtp.example.com", 587, "sender@example.com", "password",
		email.WithName("System Notification"),
		email.WithSSL(false),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Email client created: %T\n", client)

	// Output:
	// Email client created: *email.Client
}

func ExampleNewWithOptions() {
	client, err := email.NewWithOptions(
		email.WithHost("smtp.example.com"),
		email.WithPort(587),
		email.WithUsername("sender@example.com"),
		email.WithPassword("password"),
		email.WithName("System Notification"),
		email.WithSSL(false),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Email client created: %T\n", client)

	// Output:
	// Email client created: *email.Client
}

func ExampleClient_SendText() {
	fmt.Println("Send plain text email example")
	fmt.Println(`client.SendText("user@example.com", "Welcome", "Thank you for registering!")`)

	// Output:
	// Send plain text email example
	// client.SendText("user@example.com", "Welcome", "Thank you for registering!")
}

func ExampleClient_SendHTML() {
	fmt.Println("Send HTML email example")
	fmt.Println("html := `<html><body><h1>Welcome</h1></body></html>`")
	fmt.Println(`client.SendHTML("user@example.com", "Welcome", html)`)

	// Output:
	// Send HTML email example
	// html := `<html><body><h1>Welcome</h1></body></html>`
	// client.SendHTML("user@example.com", "Welcome", html)
}
