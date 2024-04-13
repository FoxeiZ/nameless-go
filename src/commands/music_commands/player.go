package music_commands

import (
	"fmt"
	"io"

	"github.com/bwmarrin/discordgo"
	"github.com/foxeiz/namelessgo/src/extractors"
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

	CurrentTrack *extractors.TrackInfo
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

func (p *Player) play(track *extractors.TrackInfo) {
	p.CurrentTrack = track

	fmt.Println(track.URL)
	defer p.afterPlayCallback(p)

	streamURL, err := track.GetStreamURL()
	if err != nil {
		p.errorCallback(p, err)
		return
	}

	encodingSession, err := dca.EncodeFile(streamURL, dca.StdEncodeOptions)
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

func (p *Player) SearchTracks(query string) ([]*extractors.TrackInfo, error) {
	data, err := extractors.Extract(query, extractors.Options{
		Playlist: true,
	})
	if err != nil {
		data, err = extractors.Extract(query, extractors.Options{
			Playlist: false,
		})
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("no data returned")
	}

	return data, nil
}

func (p *Player) AddTrack(track *extractors.TrackInfo) {
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

func (p *Player) GetTrack(index int) *extractors.TrackInfo {
	return p.queue.Peek(index)
}

func (p *Player) GetTrackList() []*extractors.TrackInfo {
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

func (p *Player) Next() *extractors.TrackInfo {
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
