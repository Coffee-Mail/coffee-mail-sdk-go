package webhook_test

import (
	"log"
	"net/http"
	"os"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/webhook"
)

func ExampleParseRequest() {
	http.HandleFunc("/webhooks/coffeemail", func(w http.ResponseWriter, r *http.Request) {
		event, err := webhook.ParseRequest(r, os.Getenv("COFFEEMAIL_WEBHOOK_SECRET"))
		if err != nil {
			http.Error(w, "assinatura inválida", http.StatusUnauthorized)
			return
		}
		log.Printf("evento %s do webhook %s", event.EventType, event.WebhookID)
		w.WriteHeader(http.StatusNoContent)
	})
}
