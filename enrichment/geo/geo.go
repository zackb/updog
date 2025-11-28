package geo

import (
	"fmt"
	"net"
	"os"

	"github.com/oschwald/geoip2-golang"
	"github.com/zackb/updog/env"
)

type Geo struct {
	cityDB *geoip2.Reader
}

type Entry struct {
	Country string
	Region  string
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
	region := ""
	if len(c.Subdivisions) > 0 {
		region = c.Subdivisions[0].IsoCode
	}
	return &Entry{
		Country: c.Country.IsoCode,
		Region:  region,
	}, nil
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
