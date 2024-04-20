package youtube

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/foxeiz/namelessgo/src/extractors"
	"github.com/tidwall/gjson"
)

var httpSession = &http.Client{}

type extractor struct{}

func init() {
	e := New()
	extractors.Register("youtube", e)
	extractors.Register("youtu", e)
	extractors.Register("", e)
}

func New() extractors.Extractor {
	return &extractor{}
}

func extractId(query string) (*string, bool, error) {
	urlParse, err := url.Parse(query)
	if err != nil {
		return nil, false, err
	}

	if urlParse.Host == "youtu.be" {
		splitPath := strings.Split(urlParse.Path, "/")

		if len(splitPath) == 2 {
			return &splitPath[1], false, nil
		}
	}

	urlQuery := urlParse.Query()

	if urlParse.Path == "/watch" {
		v := urlQuery.Get("v")
		if v != "" {
			return &v, false, nil
		}
	}

	if urlParse.Path == "/playlist" {
		list := urlQuery.Get("list")
		if list != "" {
			return &list, true, nil
		}
	}

	return nil, false, errors.New("malformed id")

}

// func extractId(url string) (*string, bool, error) {
// 	reg, _ := regexp.Compile(`(?i)(.*?)(^|playlist|watch)(?:\?)(^|\/|v=|list=)([a-z0-9_-]+)(.*)?`)

// 	match := reg.FindStringSubmatch(url)
// 	if match == nil {
// 		return nil, false, errors.New("invalid url")
// 	}

// 	if match[2] == "playlist" {
// 		if len(match[4]) == 34 {
// 			return &match[4], true, nil
// 		}
// 	}

// 	if match[2] == "watch" {
// 		if len(match[4]) == 11 {
// 			return &match[4], false, nil
// 		}
// 	}

// 	return nil, false, errors.New("malformed id")
// }

func doRequest(endpoint string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf(BaseURL, endpoint), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")

	resp, err := httpSession.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("request failed")
	}

	return resp, nil
}

func player(videoId string) (*PlayerResponse, error) {
	playerBody := PlayerBodyStruct{
		Context:      DefaultContext,
		VideoId:      videoId,
		RacyCheck:    true,
		ContentCheck: true,
	}

	playerBodyJson, _ := json.Marshal(playerBody)
	resp, err := doRequest(EndpointPlayer, playerBodyJson)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	playerResp := PlayerResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&playerResp); err != nil {
		return nil, err
	}

	return &playerResp, nil
}

func search(e *extractor, query string, param SearchParam, continuation ...string) ([]*extractors.TrackInfo, error) {
	searchBody := SearchBodyStruct{
		Context: DefaultContext,
		Query:   query,
		Params:  param,
	}

	if len(continuation) > 0 {
		searchBody.Continuation = continuation[0]
	}

	searchBodyJson, _ := json.Marshal(searchBody)
	resp, err := doRequest(EndpointSearch, searchBodyJson)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	readBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	gjsonParse := gjson.GetBytes(
		readBody,
		"contents.twoColumnSearchResultsRenderer"+
			".primaryContents.sectionListRenderer.contents"+
			".0.itemSectionRenderer.contents.#.videoRenderer",
	)
	if !gjsonParse.Exists() {
		return nil, errors.New("invalid response")
	}

	searchResult := []*extractors.TrackInfo{}
	for _, search := range gjsonParse.Array() {
		videoId := search.Get("videoId").String()
		url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoId)
		title := search.Get("title.runs.0.text").String()
		author := search.Get("ownerText.runs.0.text").String()
		authorUrl := search.Get("ownerText.runs.0.navigationEndpoint.canonicalBaseUrl").String()
		length := search.Get("lengthText.simpleText").String()
		thumbnail := search.Get("thumbnail.thumbnails.0.url").String()

		searchResult = append(searchResult, &extractors.TrackInfo{
			Site:         "youtube",
			URL:          url,
			Title:        title,
			Artist:       author,
			Duration:     time.Duration(len(length)) * time.Second,
			ThumbnailURL: &thumbnail,
			StreamData: &extractors.StreamData{
				URL: fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoId),
			},
			AuthorInfo: &extractors.AuthorInfo{
				Name: author,
				URL:  authorUrl,
			},

			IsParse:   false,
			Extractor: e,
		})
	}

	return searchResult, nil
}

func (e *extractor) Search(query string) ([]*extractors.TrackInfo, error) {
	return search(e, query, DefaultSearchParam)
}

func (e *extractor) Extract(url string, option extractors.Options) ([]*extractors.TrackInfo, error) {
	ret := make([]*extractors.TrackInfo, 0)
	youtubeId, isPlaylist, err := extractId(url)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	if !isPlaylist {
		pl, err := player(*youtubeId)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		decip, err := pl.GetBestAudio()
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		lengthSeconds, err := strconv.ParseInt(pl.VideoDetails.LengthSeconds, 10, 64)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		ret = append(ret, &extractors.TrackInfo{
			Site:         "youtube",
			URL:          url,
			Title:        pl.VideoDetails.Title,
			Artist:       pl.VideoDetails.Author,
			Duration:     time.Duration(lengthSeconds) * time.Second,
			ThumbnailURL: &pl.GetBestThumbnail().URL,
			StreamData:   decip,
			AuthorInfo: &extractors.AuthorInfo{
				Name: pl.VideoDetails.Author,
				URL:  fmt.Sprintf("https://www.youtube.com/channel/%s", pl.VideoDetails.ChannelID),
			},

			IsParse:   false,
			Extractor: e,
		})
		return ret, nil

	}
	return nil, errors.New("not implemented")
}
