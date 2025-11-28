package frontend

import (
	"github.com/zackb/updog/domain"
	"github.com/zackb/updog/user"
)

type DashboardStats struct {
	TotalPageviews int
	SelectedDomain *domain.Domain
}

type PageData struct {
	Title       string
	Description string
	Keywords    []string
	User        *user.User
	Message     string
	Domains     []*domain.Domain
	Stats       *DashboardStats
}
