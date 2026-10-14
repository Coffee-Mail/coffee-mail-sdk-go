package coffeemail_test

import (
	"context"
	"errors"
	"log"
	"time"

	coffeemail "github.com/Coffee-Mail/coffee-mail-sdk-go"
)

func ExampleNewClient() {
	client, err := coffeemail.NewClient("",
		coffeemail.WithTimeout(5*time.Second),
		coffeemail.WithLocale(coffeemail.LocaleEn),
		coffeemail.WithRetry(coffeemail.RetryPolicy{MaxAttempts: 3}),
	)
	if err != nil {
		log.Fatal(err)
	}
	_ = client
}

func ExampleEmailsService_Send() {
	client, err := coffeemail.NewClient("cm_live_...")
	if err != nil {
		log.Fatal(err)
	}
	sent, err := client.Emails.Send(context.Background(), &coffeemail.SendEmailParams{
		From:           coffeemail.Address{Email: "no-reply@suaempresa.com.br", Name: "Sua Empresa"},
		To:             []coffeemail.Address{{Email: "cliente@gmail.com"}},
		Subject:        "Seu pedido foi confirmado",
		HTML:           "<p>Obrigado pela compra!</p>",
		Tags:           []coffeemail.Tag{{Name: "fluxo", Value: "pedido"}},
		IdempotencyKey: "pedido-1234",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("email %s: %s", sent.ID, sent.Status)
}

func ExampleEmailsService_ListAll() {
	client, err := coffeemail.NewClient("cm_live_...")
	if err != nil {
		log.Fatal(err)
	}
	params := &coffeemail.ListEmailsParams{Status: coffeemail.EmailStatusBounced, Limit: 100}
	for email, err := range client.Emails.ListAll(context.Background(), params) {
		if err != nil {
			log.Fatal(err)
		}
		log.Println(email.ID, email.To)
	}
}

func ExampleError() {
	client, err := coffeemail.NewClient("cm_live_...")
	if err != nil {
		log.Fatal(err)
	}
	_, err = client.Domains.Get(context.Background(), "id-inexistente")
	if errors.Is(err, coffeemail.ErrNotFound) {
		log.Println("domínio não encontrado")
		return
	}
	var apiErr *coffeemail.Error
	if errors.As(err, &apiErr) {
		log.Printf("status %d, código %s (API: %s), requestId %s", apiErr.Status, apiErr.Code, apiErr.APICode, apiErr.RequestID)
	}
}
