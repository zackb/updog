package enrichment

import (
	"log"
	"net"
	"net/http"

	"github.com/zackb/updog/enrichment/geo"
	"github.com/zackb/updog/enrichment/ua"
)

type Enricher struct {
	g *geo.Geo
}

type Enrichment struct {
	Country    string
	Region     string
	Browser    string
	OS         string
	DeviceType string
}

func NewEnricher() (*Enricher, error) {
	g, err := geo.New()
	if err != nil {
		return nil, err
	}
	return &Enricher{g: g}, nil
}

func (e *Enricher) Enrich(req *http.Request) (*Enrichment, error) {

	res := &Enrichment{}

	ip := req.RemoteAddr

	if ip == "" || geo.IsLoopback(ip) {
		ip = req.Header.Get("X-Forwarded-For")
	}

	entry, err := e.g.Lookup(ip)
	if err != nil {
		log.Printf("geo lookup error: %v", err)
	} else {
		res.Country = entry.Country
		res.Region = entry.Region
	}
	browser, os, deviceType := ua.ParseUserAgent(req.UserAgent())
	res.Browser = browser
	res.OS = os
	res.DeviceType = deviceType

	return res, nil
}
