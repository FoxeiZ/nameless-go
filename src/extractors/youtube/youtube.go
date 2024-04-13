package youtube

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/foxeiz/namelessgo/src/extractors"
)

var httpSession = &http.Client{}

type extractor struct{}

func init() {
	e := New()
	extractors.Register("youtube", e)
	extractors.Register("youtu", e)
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

func player(videoId string) (*playerResponse, error) {
	playerBody := playerBody{
		Context:      DefaultContext,
		VideoId:      videoId,
		RacyCheck:    true,
		ContentCheck: true,
	}

	playerBodyJson, _ := json.Marshal(playerBody)
	resp, err := doRequest("player", playerBodyJson)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	playerResp := playerResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&playerResp); err != nil {
		return nil, err
	}

	return &playerResp, nil
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

			Extractor: e,
			Err:       nil,
		})
		return ret, nil

	}
	return nil, errors.New("not implemented")
}
