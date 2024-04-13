// The MIT License (MIT)

// Copyright (c) 2015 Evan Lin

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package youtube

import (
	"strconv"
	"time"

	"github.com/foxeiz/namelessgo/src/extractors"
)

const BaseURL = "https://www.youtube.com/youtubei/v1/%s?key=AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"

type Endpoint string

const (
	EndpointBrowse                    Endpoint = "browse"
	EndpointPlayer                    Endpoint = "player"
	EndpointConfig                    Endpoint = "config"
	EndpointGuide                     Endpoint = "guide"
	EndpointSearch                    Endpoint = "search"
	EndpointNext                      Endpoint = "next"
	EndpointGetTranscript             Endpoint = "get_transcript"
	EndpointMusicGetSearchSuggestions Endpoint = "music/get_search_suggestions"
	EndpointMusicGetQueue             Endpoint = "music/get_queue"
)

type ContextClient struct {
	HL            string `json:"hl"`
	GL            string `json:"gl"`
	ClientName    string `json:"clientName"`
	ClientVersion string `json:"clientVersion"`
}

type Context struct {
	Client *ContextClient `json:"client"`
}

var DefaultContext = &Context{
	Client: &ContextClient{
		HL:            "en",
		GL:            "US",
		ClientName:    "WEB",
		ClientVersion: "2.20210617.01.00",
	},
}

type playerBody struct {
	Context      *Context `json:"context"`
	VideoId      string   `json:"videoId"`
	RacyCheck    bool     `json:"racyCheckOk"`
	ContentCheck bool     `json:"contentCheckOk"`
}

// ----- the below is copied from https://github.com/kkdai/youtube/blob/master/response_data.go

type Thumbnail struct {
	URL    string
	Width  uint
	Height uint
}

type Format struct {
	Itag             int    `json:"itag"`
	URL              string `json:"url"`
	MimeType         string `json:"mimeType"`
	Quality          string `json:"quality"`
	Cipher           string `json:"signatureCipher"`
	Bitrate          int    `json:"bitrate"`
	FPS              int    `json:"fps"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
	LastModified     string `json:"lastModified"`
	ContentLength    int64  `json:"contentLength,string"`
	QualityLabel     string `json:"qualityLabel"`
	ProjectionType   string `json:"projectionType"`
	AverageBitrate   int    `json:"averageBitrate"`
	AudioQuality     string `json:"audioQuality"`
	ApproxDurationMs string `json:"approxDurationMs"`
	AudioSampleRate  string `json:"audioSampleRate"`
	AudioChannels    int    `json:"audioChannels"`

	// InitRange is only available for adaptive formats
	InitRange *struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"initRange"`

	// IndexRange is only available for adaptive formats
	IndexRange *struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"indexRange"`
}

type playerResponse struct { // trim down part that we need
	StreamingData struct {
		ExpiresInSeconds string   `json:"expiresInSeconds"`
		Formats          []Format `json:"formats"`
		AdaptiveFormats  []Format `json:"adaptiveFormats"`
		DashManifestURL  string   `json:"dashManifestUrl"`
		HlsManifestURL   string   `json:"hlsManifestUrl"`
	} `json:"streamingData"`
	VideoDetails struct {
		VideoID          string   `json:"videoId"`
		Title            string   `json:"title"`
		LengthSeconds    string   `json:"lengthSeconds"`
		Keywords         []string `json:"keywords"`
		ChannelID        string   `json:"channelId"`
		IsOwnerViewing   bool     `json:"isOwnerViewing"`
		ShortDescription string   `json:"shortDescription"`
		IsCrawlable      bool     `json:"isCrawlable"`
		Thumbnail        struct {
			Thumbnails []Thumbnail `json:"thumbnails"`
		} `json:"thumbnail"`
		AverageRating     float64 `json:"averageRating"`
		AllowRatings      bool    `json:"allowRatings"`
		ViewCount         string  `json:"viewCount"`
		Author            string  `json:"author"`
		IsPrivate         bool    `json:"isPrivate"`
		IsUnpluggedCorpus bool    `json:"isUnpluggedCorpus"`
		IsLiveContent     bool    `json:"isLiveContent"`
	} `json:"videoDetails"`

	// InitRange is only available for adaptive formats
	InitRange *struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"initRange"`

	// IndexRange is only available for adaptive formats
	IndexRange *struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"indexRange"`
}

// ----- end of copied code

func (p *playerResponse) GetBestThumbnail() *Thumbnail {
	var maxHeight uint = 0
	var thumbnail *Thumbnail

	for tIndex := range p.VideoDetails.Thumbnail.Thumbnails {
		if p.VideoDetails.Thumbnail.Thumbnails[tIndex].Height > maxHeight {
			maxHeight = p.VideoDetails.Thumbnail.Thumbnails[tIndex].Height
			thumbnail = &p.VideoDetails.Thumbnail.Thumbnails[tIndex]
		}
	}

	return thumbnail
}

func (p *playerResponse) GetBestAudio() (*extractors.StreamData, error) {
	var format Format
	var maxBitrate int = 0

	for _, f := range p.StreamingData.AdaptiveFormats {
		if f.MimeType == "audio/webm; codecs=\"opus\"" {
			if f.Bitrate > maxBitrate {
				maxBitrate = f.Bitrate
				format = f
			}
		}
	}

	var urlStream string
	var err error
	if (format.Cipher) == "" {
		urlStream, err = unThrottle(p.VideoDetails.VideoID, format.URL)
	} else {
		urlStream, err = decipherURL(p.VideoDetails.VideoID, format.Cipher)
	}
	if err != nil {
		return nil, err
	}

	expires, err := strconv.ParseInt(p.StreamingData.ExpiresInSeconds, 10, 64)
	if err != nil {
		return nil, err
	}

	return &extractors.StreamData{
			URL:     urlStream,
			Format:  format.Quality,
			Expires: time.Now().Add(time.Duration(expires) * time.Second),
		},
		nil
}
