package coffeemail

import (
	"context"
	"time"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/routes"
)

// DomainStatus é o estado de verificação de um domínio.
type DomainStatus string

// Estados de domínio.
const (
	DomainStatusPending  DomainStatus = "pending"
	DomainStatusVerified DomainStatus = "verified"
	DomainStatusFailed   DomainStatus = "failed"
)

// DomainStatusReason explica por que o domínio ainda não está verificado.
type DomainStatusReason string

// Motivos de estado de domínio.
const (
	DomainStatusReasonDNSRecordsMissing DomainStatusReason = "dns_records_missing"
	DomainStatusReasonProviderPending   DomainStatusReason = "provider_pending"
	DomainStatusReasonProviderFailed    DomainStatusReason = "provider_failed"
)

// DNSRecordType é o tipo de um registro DNS a publicar.
type DNSRecordType string

// Tipos de registro DNS.
const (
	DNSRecordTypeTXT   DNSRecordType = "TXT"
	DNSRecordTypeCNAME DNSRecordType = "CNAME"
)

// SendingProvider é o provedor que entrega os emails do domínio.
type SendingProvider string

// Provedores de envio.
const (
	SendingProviderAWSSES SendingProvider = "aws_ses"
	SendingProviderNone   SendingProvider = "none"
)

// SendingIdentityState é o estado da identidade de envio no provedor.
type SendingIdentityState string

// Estados da identidade de envio.
const (
	SendingIdentityNotProvisioned SendingIdentityState = "not_provisioned"
	SendingIdentityPending        SendingIdentityState = "pending"
	SendingIdentityReady          SendingIdentityState = "ready"
	SendingIdentityFailed         SendingIdentityState = "failed"
)

// DNSCheckResolution indica se o registro foi encontrado na consulta DNS.
type DNSCheckResolution string

// Resultados de consulta DNS.
const (
	DNSCheckFound   DNSCheckResolution = "found"
	DNSCheckAbsent  DNSCheckResolution = "absent"
	DNSCheckUnknown DNSCheckResolution = "unknown"
)

// DomainCheckName identifica uma checagem de verificação de domínio.
type DomainCheckName string

// Checagens de domínio.
const (
	DomainCheckSPF               DomainCheckName = "spf"
	DomainCheckDKIM              DomainCheckName = "dkim"
	DomainCheckDMARC             DomainCheckName = "dmarc"
	DomainCheckOwnership         DomainCheckName = "ownership"
	DomainCheckInfrastructureSPF DomainCheckName = "infrastructureSpf"
)

// DomainHealthStatus resume a saúde de autenticação do domínio.
type DomainHealthStatus string

// Estados de saúde de domínio.
const (
	DomainHealthHealthy  DomainHealthStatus = "healthy"
	DomainHealthWarning  DomainHealthStatus = "warning"
	DomainHealthCritical DomainHealthStatus = "critical"
)

// DNSRecord é um registro que deve ser publicado no DNS do domínio.
type DNSRecord struct {
	Type             DNSRecordType `json:"type"`
	Host             string        `json:"host"`
	Value            string        `json:"value"`
	TTL              *int          `json:"ttl"`
	IsInfrastructure bool          `json:"isInfrastructure"`
}

// SendingIdentity descreve a identidade de envio do domínio no provedor.
type SendingIdentity struct {
	Provider     SendingProvider      `json:"provider"`
	State        SendingIdentityState `json:"state"`
	LastSyncedAt *time.Time           `json:"lastSyncedAt"`
}

// CreateDomainResponse é o domínio criado com os registros DKIM a publicar.
type CreateDomainResponse struct {
	ID                   string       `json:"id"`
	Name                 string       `json:"name"`
	Status               DomainStatus `json:"status"`
	DKIMRecordsToPublish []DNSRecord  `json:"dkimRecordsToPublish"`
	// CreatedAt é a data de criação, quando informada pela API.
	CreatedAt *time.Time `json:"createdAt"`
}

// Domain é um domínio na listagem.
type Domain struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Status        DomainStatus        `json:"status"`
	StatusReason  *DomainStatusReason `json:"statusReason"`
	DKIMSelector  string              `json:"dkimSelector"`
	DKIMPublicKey *string             `json:"dkimPublicKey"`
	VerifiedAt    *time.Time          `json:"verifiedAt"`
	LastCheckAt   *time.Time          `json:"lastCheckAt"`
	CreatedAt     time.Time           `json:"createdAt"`
}

// DomainDetail é o detalhe de um domínio, com identidade de envio e registros DNS.
type DomainDetail struct {
	Domain
	SendingIdentity     *SendingIdentity `json:"sendingIdentity"`
	DNSRecordsToPublish []DNSRecord      `json:"dnsRecordsToPublish"`
}

// ListDomainsResponse lista os domínios da organização.
type ListDomainsResponse struct {
	Domains []Domain `json:"domains"`
}

// DNSCheck é o resultado de uma checagem de registro DNS.
type DNSCheck struct {
	OK         bool               `json:"ok"`
	Resolution DNSCheckResolution `json:"resolution"`
	Expected   string             `json:"expected"`
	Got        *string            `json:"got"`
	Matches    []string           `json:"matches"`
}

// DomainChecks reúne as checagens de verificação do domínio.
type DomainChecks struct {
	SPF               DNSCheck  `json:"spf"`
	DKIM              DNSCheck  `json:"dkim"`
	DMARC             DNSCheck  `json:"dmarc"`
	Ownership         DNSCheck  `json:"ownership"`
	InfrastructureSPF *DNSCheck `json:"infrastructureSpf"`
}

// VerifyDomainResponse é o resultado de uma verificação de domínio.
type VerifyDomainResponse struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Status          DomainStatus        `json:"status"`
	StatusReason    *DomainStatusReason `json:"statusReason"`
	VerifiedAt      *time.Time          `json:"verifiedAt"`
	Checks          DomainChecks        `json:"checks"`
	RequiredChecks  []DomainCheckName   `json:"requiredChecks"`
	SendingIdentity *SendingIdentity    `json:"sendingIdentity"`
}

// DeleteDomainResponse confirma a exclusão de um domínio.
type DeleteDomainResponse struct {
	OK bool `json:"ok"`
}

// HealthCheck é o resultado de uma checagem de saúde.
type HealthCheck struct {
	OK       bool    `json:"ok"`
	Expected string  `json:"expected"`
	Got      *string `json:"got"`
}

// DomainHealthResponse é a saúde de autenticação (SPF, DKIM, DMARC) do domínio.
type DomainHealthResponse struct {
	DomainID  string             `json:"domainId"`
	Domain    string             `json:"domain"`
	Status    DomainHealthStatus `json:"status"`
	SPF       HealthCheck        `json:"spf"`
	DKIM      HealthCheck        `json:"dkim"`
	DMARC     HealthCheck        `json:"dmarc"`
	CheckedAt time.Time          `json:"checkedAt"`
}

// DomainWarmupStatus é o progresso do aquecimento de reputação do domínio.
type DomainWarmupStatus struct {
	DomainID         string  `json:"domainId"`
	CurrentDay       int     `json:"currentDay"`
	DailyQuota       int     `json:"dailyQuota"`
	SentToday        int     `json:"sentToday"`
	RemainingToday   int     `json:"remainingToday"`
	QuotaUsedPercent int     `json:"quotaUsedPercent"`
	LastSendDate     *string `json:"lastSendDate"`
}

type createDomainBody struct {
	Name string `json:"name"`
}

// DomainsService gerencia domínios de envio.
type DomainsService struct {
	r *requester
}

// Create registra um domínio e devolve os registros DNS a publicar.
func (s *DomainsService) Create(ctx context.Context, name string) (*CreateDomainResponse, error) {
	return call[CreateDomainResponse](ctx, s.r, post(routes.Domains, createDomainBody{Name: name}))
}

// List lista os domínios da organização.
func (s *DomainsService) List(ctx context.Context) (*ListDomainsResponse, error) {
	return call[ListDomainsResponse](ctx, s.r, get(routes.Domains))
}

// Get devolve o detalhe de um domínio.
func (s *DomainsService) Get(ctx context.Context, id string) (*DomainDetail, error) {
	return call[DomainDetail](ctx, s.r, get(routes.Expand(routes.Domain, id)))
}

// Verify consulta o DNS e atualiza o estado de verificação do domínio.
func (s *DomainsService) Verify(ctx context.Context, id string) (*VerifyDomainResponse, error) {
	return call[VerifyDomainResponse](ctx, s.r, post(routes.Expand(routes.DomainVerify, id), nil))
}

// Delete exclui um domínio.
func (s *DomainsService) Delete(ctx context.Context, id string) (*DeleteDomainResponse, error) {
	return call[DeleteDomainResponse](ctx, s.r, del(routes.Expand(routes.Domain, id)))
}

// GetHealth checa SPF, DKIM e DMARC do domínio.
func (s *DomainsService) GetHealth(ctx context.Context, id string) (*DomainHealthResponse, error) {
	return call[DomainHealthResponse](ctx, s.r, post(routes.Expand(routes.DomainHealth, id), nil))
}

// GetWarmupStatus devolve o aquecimento do domínio, ou nil quando não há
// aquecimento em andamento.
func (s *DomainsService) GetWarmupStatus(ctx context.Context, id string) (*DomainWarmupStatus, error) {
	status, err := call[*DomainWarmupStatus](ctx, s.r, get(routes.Expand(routes.DomainWarmup, id)))
	if err != nil {
		return nil, err
	}
	return *status, nil
}
