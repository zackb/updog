package enrichment

import (
	"hash/crc32"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/zackb/updog/enrichment/geo"
	"github.com/zackb/updog/enrichment/ua"
	"github.com/zackb/updog/pageview"
)

type Enricher struct {
	g *geo.Geo
}

type Enrichment struct {
	Country    *pageview.Country
	Region     *pageview.Region
	City       *pageview.City
	Browser    string
	OS         string
	DeviceType string
	VisitorID  int64
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

	ip := getClientIP(req)
	userAgent := req.UserAgent()

	entry, err := e.g.Lookup(ip)
	if err != nil {
		log.Printf("geo lookup error: %v", err)
	} else {
		res.Country = entry.Country
		res.Region = entry.Region
		res.City = entry.City
	}
	browser, os, deviceType := ua.ParseUserAgent(userAgent)
	res.Browser = browser
	res.OS = os
	res.DeviceType = deviceType
	res.VisitorID = int64(crc32.ChecksumIEEE([]byte(ip + userAgent)))

	return res, nil
}

// getClientIP extracts the client's real IP address from the request
// checks X-Forwarded-For, X-Real-IP, and falls back to RemoteAddr
func getClientIP(r *http.Request) string {
	// 1. Check X-Forwarded-For (may be comma-separated)
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		ips := strings.Split(fwd, ",")
		return strings.TrimSpace(ips[0]) // first IP is the original client
	}

	// 2. Check X-Real-IP
	if real := r.Header.Get("X-Real-IP"); real != "" {
		return real
	}

	// 3. Fallback to RemoteAddr (strip port)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return host
}
