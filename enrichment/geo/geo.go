package geo

import (
	"fmt"
	"net"
	"os"

	"github.com/oschwald/geoip2-golang"
	"github.com/zackb/updog/env"
	"github.com/zackb/updog/pageview"
)

type Geo struct {
	cityDB *geoip2.Reader
}

type Entry struct {
	Country *pageview.Country
	Region  *pageview.Region
	City    *pageview.City
}

func New() (*Geo, error) {
	m := Geo{}
	err := fileExists(env.GetMaxmindCityDb())
	if err != nil {
		return nil, err
	}
	m.cityDB, err = geoip2.Open(env.GetMaxmindCityDb())
	return &m, err
}

func (m *Geo) Lookup(ip string) (*Entry, error) {
	if m.cityDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid IP address")
	}

	c, err := m.cityDB.City(parsedIP)

	if err != nil {
		return nil, err
	}

	country := &pageview.Country{
		Name: c.Country.IsoCode,
	}

	entry := &Entry{
		Country: country,
	}

	if len(c.City.Names) > 0 {
		city := &pageview.City{
			Name:       c.City.Names["en"],
			GeoNamesID: c.City.GeoNameID,
			Latitude:   c.Location.Latitude,
			Longitude:  c.Location.Longitude,
		}
		entry.City = city
	}

	if len(c.Subdivisions) > 0 {
		sub := c.Subdivisions[0]
		entry.Region = &pageview.Region{
			Name:       sub.IsoCode,
			GeoNamesID: sub.GeoNameID,
			Latitude:   c.Location.Latitude,
			Longitude:  c.Location.Longitude,
		}
	}

	return entry, nil
}

func (m *Geo) Close() error {
	if m.cityDB != nil {
		return m.cityDB.Close()
	}

	return nil
}

func fileExists(filename string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("specified data file is a directory")
	}
	return nil
}

var internal = []*net.IPNet{
	// RFC 1122, Section 3.2.1.3
	{IP: net.IPv4(127, 0, 0, 0), Mask: net.CIDRMask(8, 32)},

	// RFC 3927
	{IP: net.IPv4(169, 254, 0, 0), Mask: net.CIDRMask(16, 32)},

	// RFC 1918
	{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
	{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},
	{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)},
}

func IsLoopback(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return IsLoopbackIP(parsedIP)
}

func IsLoopbackIP(ip net.IP) bool {

	for _, n := range internal {
		if n.Contains(ip) {
			return true
		}
	}

	return false
}
