package routes

import (
	"net/url"
	"strconv"
	"strings"
)

const (
	BaseURL = "https://api.coffeemail.com.br"

	idPlaceholder        = "{id}"
	contactIDPlaceholder = "{contactId}"
	versionPlaceholder   = "{version}"
	deliveryPlaceholder  = "{deliveryId}"

	product = "/v1/product"

	APIKeyIntrospection = product + "/auth/me/api-key"

	Emails      = product + "/emails"
	EmailsBatch = Emails + "/batch"
	EmailsTags  = Emails + "/tags"
	Email       = Emails + "/" + idPlaceholder
	EmailEvents = Email + "/events"
	EmailCancel = Email + "/cancel"
	EmailResend = Email + "/resend"

	Domains              = product + "/domains"
	Domain               = Domains + "/" + idPlaceholder
	DomainVerify         = Domain + "/verify"
	DomainHealth         = Domain + "/health"
	DomainWarmup         = Domain + "/warmup"
	DomainDKIMRotate     = Domain + "/dkim/rotate"
	DomainTrackingDomain = Domain + "/tracking-domain"

	Templates           = product + "/templates"
	Template            = Templates + "/" + idPlaceholder
	TemplatesPreview    = Templates + "/preview"
	TemplatesFormat     = Templates + "/format"
	TemplatesTestRender = Templates + "/test-render"
	TemplateTestSend    = Template + "/test-send"
	TemplateVersions    = Template + "/versions"
	TemplateVersion     = TemplateVersions + "/" + versionPlaceholder
	TemplateRestore     = TemplateVersion + "/restore"

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
	WebhookDelivery     = WebhookDeliveries + "/" + deliveryPlaceholder
	WebhookReplay       = WebhookDelivery + "/replay"

	Validate = product + "/validate"

	Stats = product + "/stats"
)

func Expand(route, id string) string {
	return strings.Replace(route, idPlaceholder, url.PathEscape(id), 1)
}

func ExpandContact(route, id, contactID string) string {
	return expandChild(route, id, contactIDPlaceholder, contactID)
}

func ExpandVersion(route, id string, version int) string {
	return expandChild(route, id, versionPlaceholder, strconv.Itoa(version))
}

func ExpandDelivery(route, id, deliveryID string) string {
	return expandChild(route, id, deliveryPlaceholder, deliveryID)
}

func expandChild(route, id, placeholder, child string) string {
	return strings.Replace(Expand(route, id), placeholder, url.PathEscape(child), 1)
}
