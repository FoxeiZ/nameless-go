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

func GetExtractor(site string) Extractor {
	rwMutex.RLock()
	defer rwMutex.RUnlock()
	return extractorMap[site]
}

func Extract(Url string, option Options) ([]*TrackInfo, error) {
	u, err := url.Parse(Url)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		extractor := GetExtractor(option.SearchSite)
		return extractor.Search(Url)
	}

	domain := Domain(u.Host)
	extractor := GetExtractor(domain)
	return extractor.Extract(Url, option)
}
