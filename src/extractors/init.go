package extractors

import (
	"net/url"
	"sync"
)

var rwMutex sync.RWMutex
var extractorMap = make(map[string]Extractor)

func Register(site string, e Extractor) {
	rwMutex.Lock()
	defer rwMutex.Unlock()
	extractorMap[site] = e
}

func Extract(Url string, option Options) ([]*TrackInfo, error) {
	u, err := url.Parse(Url)
	if err != nil {
		return nil, err
	}

	domain := Domain(u.Host)
	extractor := extractorMap[domain]
	if extractor == nil {
		extractor = extractorMap[""]
	}

	videos, err := extractor.Extract(Url, option)
	if err != nil {
		return nil, err
	}

	return videos, nil
}
