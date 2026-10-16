package coffeemail

import (
	"context"
	"time"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/routes"
)

// TemplateFormat é a linguagem de origem do template.
type TemplateFormat string

// Formatos de template.
const (
	TemplateFormatHTML  TemplateFormat = "html"
	TemplateFormatReact TemplateFormat = "react"
)

// TemplateVariableType é o tipo declarado de uma variável de template.
type TemplateVariableType string

// Tipos de variável de template.
const (
	TemplateVariableString TemplateVariableType = "string"
	TemplateVariableNumber TemplateVariableType = "number"
	TemplateVariableColor  TemplateVariableType = "color"
)

// TemplateLocale é o idioma de origem do template.
type TemplateLocale string

// Idiomas de template.
const (
	TemplateLocalePtBR TemplateLocale = "pt-BR"
	TemplateLocaleEn   TemplateLocale = "en"
	TemplateLocaleEs   TemplateLocale = "es"
)

// TemplateVariable declara uma variável usada no template.
type TemplateVariable struct {
	Name          string               `json:"name"`
	Description   string               `json:"description,omitzero"`
	Type          TemplateVariableType `json:"type,omitzero"`
	FallbackValue string               `json:"fallbackValue,omitzero"`
}

// CreateTemplateParams são os dados de um template novo.
type CreateTemplateParams struct {
	Name         string             `json:"name"`
	HTML         string             `json:"html"`
	Alias        string             `json:"alias,omitzero"`
	Subject      string             `json:"subject,omitzero"`
	Preheader    string             `json:"preheader,omitzero"`
	ContentJSON  map[string]any     `json:"contentJson,omitzero"`
	TextPayload  string             `json:"textPayload,omitzero"`
	Variables    []TemplateVariable `json:"variables,omitzero"`
	Format       TemplateFormat     `json:"format,omitzero"`
	SourceLocale TemplateLocale     `json:"sourceLocale,omitzero"`
}

// UpdateTemplateParams altera um template. Campos nil não são alterados.
type UpdateTemplateParams struct {
	Name         *string            `json:"name,omitzero"`
	HTML         *string            `json:"html,omitzero"`
	Alias        *string            `json:"alias,omitzero"`
	Subject      *string            `json:"subject,omitzero"`
	Preheader    *string            `json:"preheader,omitzero"`
	ContentJSON  map[string]any     `json:"contentJson,omitzero"`
	TextPayload  *string            `json:"textPayload,omitzero"`
	Variables    []TemplateVariable `json:"variables,omitzero"`
	Format       TemplateFormat     `json:"format,omitzero"`
	IsActive     *bool              `json:"isActive,omitzero"`
	SourceLocale TemplateLocale     `json:"sourceLocale,omitzero"`
}

// Template é um template de email.
type Template struct {
	ID             string             `json:"id"`
	OrganisationID string             `json:"organisationId"`
	Name           string             `json:"name"`
	Alias          *string            `json:"alias"`
	Subject        *string            `json:"subject"`
	Preheader      *string            `json:"preheader"`
	HTML           string             `json:"html"`
	ContentJSON    map[string]any     `json:"contentJson"`
	TextPayload    *string            `json:"textPayload"`
	Variables      []TemplateVariable `json:"variables"`
	IsActive       bool               `json:"isActive"`
	Format         TemplateFormat     `json:"format"`
	SourceLocale   TemplateLocale     `json:"sourceLocale"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

// ListTemplatesResponse lista os templates.
type ListTemplatesResponse struct {
	Data []Template `json:"data"`
}

// PreviewTemplateParams renderiza um HTML de template sem salvá-lo.
type PreviewTemplateParams struct {
	HTML      string         `json:"html"`
	Format    TemplateFormat `json:"format,omitzero"`
	Variables map[string]any `json:"variables,omitzero"`
}

// PreviewTemplateResponse é o template renderizado.
type PreviewTemplateResponse struct {
	HTML string `json:"html"`
	Text string `json:"text"`
}

// FormatTemplateParams formata o código de um template.
type FormatTemplateParams struct {
	HTML   string         `json:"html"`
	Format TemplateFormat `json:"format,omitzero"`
}

// FormatTemplateResponse é o código formatado.
type FormatTemplateResponse struct {
	HTML string `json:"html"`
}

// TestRenderTemplateParams renderiza um template e relata a sanitização.
type TestRenderTemplateParams struct {
	HTML      string         `json:"html"`
	Format    TemplateFormat `json:"format,omitzero"`
	Variables map[string]any `json:"variables,omitzero"`
}

// SanitizeReport conta os elementos removidos pela sanitização.
type SanitizeReport struct {
	Scripts         int `json:"scripts"`
	Iframes         int `json:"iframes"`
	JavascriptHrefs int `json:"javascriptHrefs"`
	EventHandlers   int `json:"eventHandlers"`
}

// TestRenderTemplateResponse é o template renderizado com o relatório de sanitização.
type TestRenderTemplateResponse struct {
	HTML           string         `json:"html"`
	Text           string         `json:"text"`
	SanitizeReport SanitizeReport `json:"sanitizeReport"`
}

// TestSendTemplateParams envia o template para um destinatário de teste.
type TestSendTemplateParams struct {
	To        string         `json:"to"`
	From      string         `json:"from,omitzero"`
	Subject   string         `json:"subject,omitzero"`
	Variables map[string]any `json:"variables,omitzero"`
}

// TestSendResponse é o resultado de um envio de teste.
type TestSendResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"messageId"`
	Recipient string `json:"recipient"`
}

// TemplateVersionSummary é uma versão gravada de um template, sem o conteúdo.
type TemplateVersionSummary struct {
	// Version é o número da versão, crescente a partir de 1.
	Version int     `json:"version"`
	Name    string  `json:"name"`
	Subject *string `json:"subject"`
	// CreatedBy é o usuário que gerou a versão, ou nil quando veio de chave de API.
	CreatedBy *string   `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
}

// TemplateVersion é uma versão gravada de um template, com o conteúdo completo.
type TemplateVersion struct {
	TemplateVersionSummary
	TemplateID   string             `json:"templateId"`
	Preheader    *string            `json:"preheader"`
	HTML         string             `json:"html"`
	ContentJSON  map[string]any     `json:"contentJson"`
	TextPayload  *string            `json:"textPayload"`
	Variables    []TemplateVariable `json:"variables"`
	Format       TemplateFormat     `json:"format"`
	SourceLocale TemplateLocale     `json:"sourceLocale"`
}

// ListTemplateVersionsResponse é uma página de versões, da mais recente para a
// mais antiga.
type ListTemplateVersionsResponse struct {
	Data  []TemplateVersionSummary `json:"data"`
	Total int                      `json:"total"`
}

// TemplatesService gerencia templates de email.
type TemplatesService struct {
	r *requester
}

// Create cria um template.
func (s *TemplatesService) Create(ctx context.Context, params *CreateTemplateParams) (*Template, error) {
	return call[Template](ctx, s.r, post(routes.Templates, derefOrZero(params)))
}

// List lista os templates.
func (s *TemplatesService) List(ctx context.Context) (*ListTemplatesResponse, error) {
	return call[ListTemplatesResponse](ctx, s.r, get(routes.Templates))
}

// Get devolve um template.
func (s *TemplatesService) Get(ctx context.Context, id string) (*Template, error) {
	return call[Template](ctx, s.r, get(routes.Expand(routes.Template, id)))
}

// Update altera um template.
func (s *TemplatesService) Update(ctx context.Context, id string, params *UpdateTemplateParams) (*Template, error) {
	return call[Template](ctx, s.r, patch(routes.Expand(routes.Template, id), derefOrZero(params)))
}

// Delete exclui um template.
func (s *TemplatesService) Delete(ctx context.Context, id string) error {
	return s.r.do(ctx, del(routes.Expand(routes.Template, id)), nil)
}

// Preview renderiza um HTML de template com variáveis, sem salvá-lo.
func (s *TemplatesService) Preview(ctx context.Context, params *PreviewTemplateParams) (*PreviewTemplateResponse, error) {
	return call[PreviewTemplateResponse](ctx, s.r, post(routes.TemplatesPreview, derefOrZero(params)))
}

// PreviewByID busca o template salvo e o renderiza com variables, que pode ser nil.
func (s *TemplatesService) PreviewByID(ctx context.Context, id string, variables map[string]any) (*PreviewTemplateResponse, error) {
	template, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.Preview(ctx, &PreviewTemplateParams{HTML: template.HTML, Format: template.Format, Variables: variables})
}

// Format formata o código de um template.
func (s *TemplatesService) Format(ctx context.Context, params *FormatTemplateParams) (*FormatTemplateResponse, error) {
	return call[FormatTemplateResponse](ctx, s.r, post(routes.TemplatesFormat, derefOrZero(params)))
}

// TestRender renderiza o template e informa o que a sanitização removeu.
func (s *TemplatesService) TestRender(ctx context.Context, params *TestRenderTemplateParams) (*TestRenderTemplateResponse, error) {
	return call[TestRenderTemplateResponse](ctx, s.r, post(routes.TemplatesTestRender, derefOrZero(params)))
}

// TestSend envia o template salvo para um destinatário de teste.
func (s *TemplatesService) TestSend(ctx context.Context, id string, params *TestSendTemplateParams) (*TestSendResponse, error) {
	return call[TestSendResponse](ctx, s.r, post(routes.Expand(routes.TemplateTestSend, id), derefOrZero(params)))
}

// ListVersions lista as versões gravadas do template, da mais recente para a
// mais antiga. params pode ser nil; Limit vai de 1 a 100 (padrão da API: 20).
func (s *TemplatesService) ListVersions(ctx context.Context, id string, params *OffsetParams) (*ListTemplateVersionsResponse, error) {
	return call[ListTemplateVersionsResponse](ctx, s.r, get(routes.Expand(routes.TemplateVersions, id)).withQuery(params.values()))
}

// GetVersion devolve uma versão do template com o conteúdo completo. Versão
// inexistente gera [ErrNotFound] com APICode TEMPLATE_VERSION_NOT_FOUND.
func (s *TemplatesService) GetVersion(ctx context.Context, id string, version int) (*TemplateVersion, error) {
	return call[TemplateVersion](ctx, s.r, get(routes.ExpandVersion(routes.TemplateVersion, id, version)))
}

// RestoreVersion restaura o conteúdo de uma versão no template e devolve o
// template atualizado. Versão inexistente gera [ErrNotFound] com APICode
// TEMPLATE_VERSION_NOT_FOUND.
func (s *TemplatesService) RestoreVersion(ctx context.Context, id string, version int) (*Template, error) {
	return call[Template](ctx, s.r, post(routes.ExpandVersion(routes.TemplateRestore, id, version), nil))
}
