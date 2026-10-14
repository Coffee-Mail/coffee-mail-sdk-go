// Command send envia um email de teste em modo sandbox. Defina COFFEEMAIL_API_KEY,
// COFFEEMAIL_FROM e COFFEEMAIL_TO e rode: go run ./examples/send
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	coffeemail "github.com/Coffee-Mail/coffee-mail-sdk-go"
)

const exitFailure = 1

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFailure)
	}
}

func run() error {
	client, err := coffeemail.NewClient("")
	if err != nil {
		return err
	}
	sent, err := client.Emails.Send(context.Background(), &coffeemail.SendEmailParams{
		From:    coffeemail.Address{Email: os.Getenv("COFFEEMAIL_FROM")},
		To:      []coffeemail.Address{{Email: os.Getenv("COFFEEMAIL_TO")}},
		Subject: "Teste do SDK Go",
		Text:    "Enviado pelo exemplo do coffee-mail-sdk-go.",
		Sandbox: true,
	})
	var apiErr *coffeemail.Error
	if errors.As(err, &apiErr) {
		return fmt.Errorf("falha %s (status %d, requestId %q): %s", apiErr.Code, apiErr.Status, apiErr.RequestID, apiErr.Message)
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(os.Stdout, "email %s enfileirado com status %s\n", sent.ID, sent.Status)
	return err
}
