package env

import (
	"log"
	"os"
	"strconv"
)

const (
	EnvHttpPort      = "HTTP_PORT"
	EnvDsn           = "DATABASE_URL"
	EnvMaxmindCityDb = "MAXMIND_CITY_DB"
)

var ecache = map[string]string{}

func GetString(name, def string) string {

	s := ecache[name]
	if s == "" {
		s = os.Getenv(name)
	}

	if s == "" {
		s = def
	}
	return s
}

func GetInt(name string, def int) int {
	s := os.Getenv(name)
	if s == "" {
		return def
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("ERROR: failed parsing env %s %s %s\n", name, s, err.Error())
	}
	return i
}

func IsDev() bool {
	s := GetString("DEV", "false")
	if s == "true" || s == "1" {
		return true
	}
	return false
}

func GetHTTPPort() int {
	return GetInt(EnvHttpPort, 8081)
}

func GetDsn() string {
	return GetString(EnvDsn, "")
}

func GetMaxmindCityDb() string {
	return GetString(EnvMaxmindCityDb, "data/maxmind/GeoLite2-City.mmdb")
}
