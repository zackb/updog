package frontend

import (
	"github.com/zackb/updog/domain"
	"github.com/zackb/updog/pageview"
	"github.com/zackb/updog/user"
)

type DashboardStats struct {
	TotalPageviews int
	SelectedDomain *domain.Domain
	Aggregated     *pageview.AggregatedStats
	// TODO: remove in favor of htmx loaded components
	GraphData   []*pageview.AggregatedPoint
	MaxViews    int64
	TopPages    []*pageview.PageStats
	DeviceUsage []*pageview.DeviceStats
}

type PageData struct {
	Title       string
	Description string
	Keywords    []string
	User        *user.User
	Message     string
	Domains     []*domain.Domain
	Stats       *DashboardStats
	Slug        string
	Error       string
	Data        map[string]any
}
