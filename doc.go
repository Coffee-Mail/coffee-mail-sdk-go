// Package coffeemail é o SDK oficial em Go da API CoffeeMail, de envio de emails
// transacionais e broadcasts.
//
// Crie um [Client] com [NewClient] e use os serviços expostos como campos:
//
//	client, err := coffeemail.NewClient("cm_live_...")
//	if err != nil {
//		return err
//	}
//	sent, err := client.Emails.Send(ctx, &coffeemail.SendEmailParams{
//		From:    coffeemail.Address{Email: "no-reply@suaempresa.com.br"},
//		To:      []coffeemail.Address{{Email: "cliente@gmail.com"}},
//		Subject: "Bem-vindo",
//		HTML:    "<p>Olá!</p>",
//	})
//
// Todo método recebe um [context.Context] e devolve o dado ou um erro do tipo
// [*Error]. Use [errors.Is] com as sentinelas ([ErrNotFound], [ErrRateLimited], ...)
// ou [errors.As] para ler status, código e requestId.
//
// A verificação de assinatura de webhooks fica no subpacote webhook e não depende
// do Client.
package coffeemail
