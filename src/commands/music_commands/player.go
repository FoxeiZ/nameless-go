package music_commands

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	extractors "github.com/iawia002/lux/extractors"
	_ "github.com/iawia002/lux/extractors/douyin"
	_ "github.com/iawia002/lux/extractors/douyu"
	_ "github.com/iawia002/lux/extractors/facebook"
	_ "github.com/iawia002/lux/extractors/geekbang"
	_ "github.com/iawia002/lux/extractors/haokan"
	_ "github.com/iawia002/lux/extractors/hupu"
	_ "github.com/iawia002/lux/extractors/huya"
	_ "github.com/iawia002/lux/extractors/instagram"
	_ "github.com/iawia002/lux/extractors/kuaishou"
	_ "github.com/iawia002/lux/extractors/mgtv"
	_ "github.com/iawia002/lux/extractors/miaopai"
	_ "github.com/iawia002/lux/extractors/netease"
	_ "github.com/iawia002/lux/extractors/pinterest"
	_ "github.com/iawia002/lux/extractors/qq"
	_ "github.com/iawia002/lux/extractors/reddit"
	_ "github.com/iawia002/lux/extractors/rumble"
	_ "github.com/iawia002/lux/extractors/tangdou"
	_ "github.com/iawia002/lux/extractors/tiktok"
	_ "github.com/iawia002/lux/extractors/twitter"
	_ "github.com/iawia002/lux/extractors/udn"
	_ "github.com/iawia002/lux/extractors/universal"
	_ "github.com/iawia002/lux/extractors/vimeo"
	_ "github.com/iawia002/lux/extractors/vk"
	_ "github.com/iawia002/lux/extractors/weibo"
	_ "github.com/iawia002/lux/extractors/xiaohongshu"
	_ "github.com/iawia002/lux/extractors/xvideos"
	_ "github.com/iawia002/lux/extractors/yinyuetai"
	_ "github.com/iawia002/lux/extractors/youku"
	_ "github.com/iawia002/lux/extractors/youtube"
	_ "github.com/iawia002/lux/extractors/zhihu"
	"github.com/pkg/errors"

	"github.com/FoxeiZ/dca"
)

type Player struct {
	channelID       string
	voiceConnection *discordgo.VoiceConnection

	queue     *Queue
	position  int
	isPlaying bool

	errorCallback     func(p *Player, err error)
	afterPlayCallback func(p *Player)

	CurrentTrack *TrackInfo
}

func NewPlayer(
	channelID string,
	voiceConnection *discordgo.VoiceConnection,
	errorCallback func(p *Player, err error),
	afterPlayCallback func(p *Player),
) *Player {
	return &Player{
		channelID:       channelID,
		voiceConnection: voiceConnection,
		queue:           NewQueue(),
		position:        0,
		isPlaying:       false,

		errorCallback:     errorCallback,
		afterPlayCallback: afterPlayCallback,
	}
}

func sortedStreams(streams map[string]*extractors.Stream) []*extractors.Stream {
	sortedStreams := make([]*extractors.Stream, 0, len(streams))
	for _, data := range streams {
		sortedStreams = append(sortedStreams, data)
	}
	if len(sortedStreams) > 1 {
		sort.SliceStable(
			sortedStreams, func(i, j int) bool { return sortedStreams[i].Size > sortedStreams[j].Size },
		)
	}
	return sortedStreams
}

func findAudioOnlyStream(streams map[string]*extractors.Stream, formats *string) (*string, error) {
	reg, err := regexp.Compile("audio+")
	if err != nil {
		return nil, err
	}

	if formats != nil {
		for _, f := range strings.Split(*formats, ",") {
			stream, ok := streams[f]
			if ok && reg.MatchString(stream.Quality) {
				return &stream.Parts[0].URL, nil
			}
		}
	}

	for _, s := range sortedStreams(streams) {
		// Looking for the best quality
		if reg.MatchString(s.Quality) {
			return &s.Parts[0].URL, nil
		}
	}

	return nil, errors.Errorf("no audio stream found")
}

func buildStreams(data []*extractors.Data) []*TrackInfo {
	trackList := make([]*TrackInfo, 0, len(data))
	for _, d := range data {
		url, err := findAudioOnlyStream(d.Streams, nil)
		if err != nil {
			log.Println(err)
			continue
		}
		trackList = append(trackList, &TrackInfo{
			Title:  d.Title,
			Artist: "N/A",
			URI:    d.URL,
			SITE:   d.Site,
			URL:    *url,
		})
	}
	return trackList
}

func (p *Player) play(track *TrackInfo) {
	p.CurrentTrack = track

	fmt.Println(track.URL)
	defer p.afterPlayCallback(p)

	encodingSession, err := dca.EncodeFile(track.URL, dca.StdEncodeOptions)
	if err != nil {
		p.errorCallback(p, err)
		return
	}
	defer encodingSession.Cleanup()

	done := make(chan error)
	dca.NewStream(encodingSession, p.voiceConnection, done)
	err = <-done
	if err != nil && err != io.EOF {
		p.errorCallback(p, err)
		return
	}
}

func (p *Player) SearchTracks(query string) ([]*TrackInfo, error) {
	data, err := extractors.Extract(query, extractors.Options{
		Playlist:     true,
		ThreadNumber: 4,
	})
	if err != nil {
		data, err = extractors.Extract(query, extractors.Options{
			Playlist: false,
		})
	}
	if err != nil {
		return nil, err
	}

	return buildStreams(data), nil
}

func (p *Player) AddTrack(track *TrackInfo) {
	playAfter := false
	if !p.isPlaying && len(p.queue.TrackList) == 0 {
		playAfter = true
	}

	p.queue.Enqueue(track)

	if playAfter {
		p.Next()
	}
}

func (p *Player) RemoveTrack(index int) {
	p.queue.Pop(index)
}

func (p *Player) GetTrack(index int) *TrackInfo {
	return p.queue.Peek(index)
}

func (p *Player) GetTrackList() []*TrackInfo {
	return p.queue.GetTrackList()
}

func (p *Player) GetPosition() int {
	return p.position
}

func (p *Player) Play() {
	p.isPlaying = true
	p.play(p.CurrentTrack)
}

func (p *Player) Pause() {
	p.isPlaying = false
}

func (p *Player) Stop() {
	p.isPlaying = false
}

func (p *Player) Next() *TrackInfo {
	p.CurrentTrack = p.queue.Dequeue()
	if p.CurrentTrack == nil {
		p.isPlaying = false
		return nil
	}

	p.Play()
	return p.CurrentTrack
}

func (p *Player) SetErrorCallback(callback func(p *Player, err error)) {
	p.errorCallback = callback
}

func (p *Player) SetafterPlayCallback(callback func(p *Player)) {
	p.afterPlayCallback = callback
}
