package player

import (
	"io"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/foxeiz/namelessgo/src/extractors"
	"github.com/pkg/errors"

	"github.com/FoxeiZ/dca"
)

type Player struct {
	sync.Mutex
	channelID       string
	voiceConnection *discordgo.VoiceConnection

	queue     *Queue
	position  int
	isPlaying bool

	errorCallback     func(p *Player, err error)
	afterPlayCallback func(p *Player)

	CurrentTrack *extractors.TrackInfo

	currentStreamingSession *dca.StreamingSession
	currentEncodingSession  *dca.EncodeSession
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

		queue:     NewQueue(),
		position:  0,
		isPlaying: false,

		errorCallback:     errorCallback,
		afterPlayCallback: afterPlayCallback,
	}
}

func (p *Player) cleanup() {
	p.Lock()
	defer p.Unlock()

	p.isPlaying = false
	p.CurrentTrack = nil
	p.currentStreamingSession = nil
	p.currentEncodingSession = nil
}

func (p *Player) Cleanup() {
	p.cleanup()
	p.voiceConnection.Disconnect()
}

func (p *Player) play(track *extractors.TrackInfo) {
	p.CurrentTrack = track

	defer p.afterPlayCallback(p)

	streamURL, err := track.GetStreamURL()
	if err != nil {
		p.errorCallback(p, err)
		return
	}

	p.currentEncodingSession, err = dca.EncodeFile(streamURL, dca.StdEncodeOptions)
	if err != nil {
		p.errorCallback(p, err)
		return
	}
	defer p.currentEncodingSession.Cleanup()
	defer p.cleanup()

	done := make(chan error)
	p.currentStreamingSession = dca.NewStream(p.currentEncodingSession, p.voiceConnection, done)

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

func (p *Player) RemoveTrack(index int) error {
	p.Lock()
	defer p.Unlock()

	if index < 0 || index >= len(p.queue.TrackList) {
		return errors.New("invalid index")
	}

	p.queue.Pop(index)

	return nil
}

func (p *Player) GetChannelID() string {
	return p.channelID
}

func (p *Player) GetTrack(index int) *extractors.TrackInfo {
	p.Lock()
	defer p.Unlock()
	return p.queue.Peek(index)
}

func (p *Player) GetTrackList() []*extractors.TrackInfo {
	p.Lock()
	defer p.Unlock()
	return p.queue.GetTrackList()
}

func (p *Player) GetPlaybackPosition() time.Duration {
	if p.currentStreamingSession == nil {
		return 0
	}
	return p.currentStreamingSession.PlaybackPosition()
}

func (p *Player) Play() {
	p.Lock()
	defer p.Unlock()

	if p.currentStreamingSession != nil {
		p.isPlaying = true
		p.currentStreamingSession.SetPaused(false)
		return
	}

	p.isPlaying = true
	p.play(p.CurrentTrack)
}

func (p *Player) Pause() {
	p.Lock()
	defer p.Unlock()

	if p.currentStreamingSession == nil {
		return
	}

	p.currentStreamingSession.SetPaused(true)
	p.isPlaying = false
}

func (p *Player) Stop() {
	p.Lock()
	defer p.Unlock()

	if p.currentStreamingSession == nil && p.currentEncodingSession == nil {
		return
	}

	p.isPlaying = false
	p.currentEncodingSession.Stop()
	p.cleanup()
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
