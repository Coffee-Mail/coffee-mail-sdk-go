package coffeemail

import (
	"context"

	"github.com/Coffee-Mail/coffee-mail-sdk-go/internal/routes"
)

// StatsGranularity é o agrupamento temporal da série.
type StatsGranularity string

// Granularidades de métricas.
const (
	StatsGranularityDay   StatsGranularity = "day"
	StatsGranularityWeek  StatsGranularity = "week"
	StatsGranularityMonth StatsGranularity = "month"
)

// StatsPeriod é um período predefinido.
type StatsPeriod string

// Períodos de métricas.
const (
	StatsPeriodLast7Days  StatsPeriod = "last7d"
	StatsPeriodLast30Days StatsPeriod = "last30d"
	StatsPeriodLast90Days StatsPeriod = "last90d"
)

// StatsEmailType filtra as métricas por tipo de envio.
type StatsEmailType string

// Tipos de envio para métricas.
const (
	StatsEmailTypeAll           StatsEmailType = "all"
	StatsEmailTypeTransactional StatsEmailType = "transactional"
	StatsEmailTypeBroadcast     StatsEmailType = "broadcast"
)

// GetStatsParams filtra as métricas. Todos os campos são opcionais.
type GetStatsParams struct {
	DomainID string
	Tag      string
	// StartDate e EndDate seguem o formato AAAA-MM-DD.
	StartDate   string
	EndDate     string
	Granularity StatsGranularity
	Period      StatsPeriod
	EmailType   StatsEmailType
}

// StatsBreakdownItem são os totais de um tipo de envio.
type StatsBreakdownItem struct {
	Sent      int `json:"sent"`
	Delivered int `json:"delivered"`
	Bounced   int `json:"bounced"`
	Opened    int `json:"opened"`
	Clicked   int `json:"clicked"`
}

// StatsBreakdown separa os totais por tipo de envio.
type StatsBreakdown struct {
	Transactional StatsBreakdownItem `json:"transactional"`
	Broadcast     StatsBreakdownItem `json:"broadcast"`
}

// StatsClient é um cliente de email entre os que mais abriram mensagens.
type StatsClient struct {
	Name       string  `json:"name"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
	DeviceType string  `json:"deviceType"`
}

// StatsPoint é um ponto da série temporal.
type StatsPoint struct {
	Period        string `json:"period"`
	Sent          int    `json:"sent"`
	Delivered     int    `json:"delivered"`
	Bounced       int    `json:"bounced"`
	Failed        int    `json:"failed"`
	Opened        int    `json:"opened"`
	Clicked       int    `json:"clicked"`
	Transactional int    `json:"transactional"`
	Broadcast     int    `json:"broadcast"`
}

// StatsResponse são as métricas de envio. Taxas vão de 0 a 100.
type StatsResponse struct {
	TotalSent      int             `json:"totalSent"`
	TotalDelivered int             `json:"totalDelivered"`
	TotalBounced   int             `json:"totalBounced"`
	TotalFailed    int             `json:"totalFailed"`
	TotalOpened    int             `json:"totalOpened"`
	TotalClicked   int             `json:"totalClicked"`
	DeliveryRate   float64         `json:"deliveryRate"`
	OpenRate       float64         `json:"openRate"`
	ClickRate      float64         `json:"clickRate"`
	Breakdown      *StatsBreakdown `json:"breakdown"`
	TopClients     []StatsClient   `json:"topClients"`
	Data           []StatsPoint    `json:"data"`
}

// StatsService consulta métricas de envio.
type StatsService struct {
	r *requester
}

// Get devolve as métricas de envio. params pode ser nil.
func (s *StatsService) Get(ctx context.Context, params *GetStatsParams) (*StatsResponse, error) {
	v := derefOrZero(params)
	query := newQuery().
		str("domainId", v.DomainID).
		str("tag", v.Tag).
		str("startDate", v.StartDate).
		str("endDate", v.EndDate).
		str("granularity", string(v.Granularity)).
		str("period", string(v.Period)).
		str("emailType", string(v.EmailType)).
		values()
	return call[StatsResponse](ctx, s.r, get(routes.Stats).withQuery(query))
}
