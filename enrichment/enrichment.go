package enrichment

import (
	"log"
	"net/http"

	"github.com/zackb/updog/enrichment/geo"
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

	entry, err := e.g.Lookup(req.RemoteAddr)
	if err != nil {
		log.Printf("geo lookup error: %v", err)
	} else {
		res.Country = entry.Country
		res.Region = entry.Region
	}
	browser, os, deviceType := parseUserAgent(req.UserAgent())
	res.Browser = browser
	res.OS = os
	res.DeviceType = deviceType

	return res, nil
}

func parseUserAgent(ua string) (browser string, os string, deviceType string) {
	return "Unknown", "Unknown", "Unknown"
}
