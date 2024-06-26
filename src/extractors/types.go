package extractors

import (
	"time"
)

type AuthorInfo struct {
	Name string
	URL  string
}

type StreamData struct {
	// URL of the stream
	URL string
	// Format of the stream (ex: "mp3")
	Format string
	// Expiration time of the stream
	Expires time.Time
}

type TrackInfo struct {
	Site         string
	URL          string
	Title        string
	Artist       string
	Duration     time.Duration
	ThumbnailURL *string
	StreamData   *StreamData
	AuthorInfo   *AuthorInfo

	// IsParse indicates if the track info has been parsed, return `false` if the track info is return from a search
	IsParse bool

	Extractor Extractor
}

func (t *TrackInfo) Update(newT *TrackInfo) {
	t.Site = newT.Site
	t.URL = newT.URL
	t.Title = newT.Title
	t.Artist = newT.Artist
	t.Duration = newT.Duration
	t.ThumbnailURL = newT.ThumbnailURL
	t.StreamData = newT.StreamData
	t.AuthorInfo = newT.AuthorInfo
	t.Extractor = newT.Extractor
}

func (t *TrackInfo) GetStreamURL() (string, error) {
	if t.IsParse && t.StreamData != nil {
		if !t.StreamData.Expires.Before(time.Now()) {
			return t.StreamData.URL, nil
		}
	}

	reExtract, err := t.Extractor.Extract(t.URL, Options{})
	if err != nil {
		return "", err
	}

	t.Update(reExtract[0])
	return t.StreamData.URL, nil
}

type Options struct {
	// Playlist indicates if we need to extract the whole playlist rather than the single video.
	Playlist bool
	// Items defines wanted items from a playlist. Separated by commas like: 1,5,6,8-10.
	Items string
	// ItemStart defines the starting item of a playlist.
	ItemStart int
	// ItemEnd defines the ending item of a playlist.
	ItemEnd int

	// Which site to search
	SearchSite string

	// ThreadNumber defines how many threads will use in the extraction, only works when Playlist is true.
	ThreadNumber int
	Cookie       string
}

type Extractor interface {
	// Extract is the main function to extract the data.
	Extract(url string, option Options) ([]*TrackInfo, error)

	Search(query string) ([]*TrackInfo, error)
}
