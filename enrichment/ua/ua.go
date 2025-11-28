package ua

import (
	"github.com/mileusna/useragent"
)

func ParseUserAgent(ua string) (browser string, os string, deviceType string) {
	uaParsed := useragent.Parse(ua)

	browser = uaParsed.Name + " " + uaParsed.Version
	os = uaParsed.OS

	if uaParsed.Mobile {
		deviceType = "Mobile"
	} else if uaParsed.Tablet {
		deviceType = "Tablet"
	} else {
		deviceType = "Desktop"
	}

	return browser, os, deviceType
}
