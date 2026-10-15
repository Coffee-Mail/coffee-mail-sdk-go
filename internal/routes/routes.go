package routes

import (
	"net/url"
	"strings"
)

const (
	BaseURL = "https://api.coffeemail.com.br"

	idPlaceholder        = "{id}"
	contactIDPlaceholder = "{contactId}"

	product = "/v1/product"

	APIKeyIntrospection = product + "/auth/me/api-key"

	Emails      = product + "/emails"
	EmailsBatch = Emails + "/batch"
	EmailsTags  = Emails + "/tags"
	Email       = Emails + "/" + idPlaceholder
	EmailEvents = Email + "/events"
	EmailCancel = Email + "/cancel"
	EmailResend = Email + "/resend"

	Domains      = product + "/domains"
	Domain       = Domains + "/" + idPlaceholder
	DomainVerify = Domain + "/verify"
	DomainHealth = Domain + "/health"
	DomainWarmup = Domain + "/warmup"

	Templates           = product + "/templates"
	Template            = Templates + "/" + idPlaceholder
	TemplatesPreview    = Templates + "/preview"
	TemplatesFormat     = Templates + "/format"
	TemplatesTestRender = Templates + "/test-render"
	TemplateTestSend    = Template + "/test-send"

	Audiences            = product + "/audiences"
	Audience             = Audiences + "/" + idPlaceholder
	AudienceContacts     = Audience + "/contacts"
	AudienceContactsBulk = AudienceContacts + "/bulk"
	AudienceContact      = AudienceContacts + "/" + contactIDPlaceholder

	Broadcasts        = product + "/broadcasts"
	Broadcast         = Broadcasts + "/" + idPlaceholder
	BroadcastSend     = Broadcast + "/send"
	BroadcastTestSend = Broadcast + "/test-send"
	BroadcastCancel   = Broadcast + "/cancel"

	Senders       = product + "/senders"
	SendersVerify = Senders + "/verify"
	Sender        = Senders + "/" + idPlaceholder
	SenderResend  = Sender + "/resend-verification"

	Suppressions          = product + "/suppressions"
	SuppressionsBulk      = Suppressions + "/bulk"
	Suppression           = Suppressions + "/" + idPlaceholder
	SuppressionReactivate = Suppression + "/reactivate"

	Webhooks            = product + "/webhooks"
	Webhook             = Webhooks + "/" + idPlaceholder
	WebhookTest         = Webhook + "/test"
	WebhookDeliveries   = Webhook + "/deliveries"
	WebhookRotateSecret = Webhook + "/rotate-secret"

	Stats = product + "/stats"
)

func Expand(route, id string) string {
	return strings.Replace(route, idPlaceholder, url.PathEscape(id), 1)
}

func ExpandContact(route, id, contactID string) string {
	return strings.Replace(Expand(route, id), contactIDPlaceholder, url.PathEscape(contactID), 1)
}
