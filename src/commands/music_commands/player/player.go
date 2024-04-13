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
	channelID       string
	voiceConnection *discordgo.VoiceConnection

	queue     *Queue
	lock      *sync.Mutex
	position  int
	isPlaying bool

	errorCallback     func(p *Player, err error)
	afterPlayCallback func(p *Player)

	CurrentTrack            *extractors.TrackInfo
	CurrentStreamingSession *dca.StreamingSession
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

		lock:      &sync.Mutex{},
		queue:     NewQueue(),
		position:  0,
		isPlaying: false,

		errorCallback:     errorCallback,
		afterPlayCallback: afterPlayCallback,
	}
}

func (p *Player) cleanup() {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.isPlaying = false
	p.CurrentTrack = nil
	p.CurrentStreamingSession = nil
}

func (p *Player) play(track *extractors.TrackInfo) {
	p.CurrentTrack = track

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
	defer p.cleanup()

	done := make(chan error)
	p.CurrentStreamingSession = dca.NewStream(encodingSession, p.voiceConnection, done)

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
	p.lock.Lock()
	defer p.lock.Unlock()

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
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.queue.Peek(index)
}

func (p *Player) GetTrackList() []*extractors.TrackInfo {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.queue.GetTrackList()
}

func (p *Player) GetPlaybackPosition() time.Duration {
	if p.CurrentStreamingSession == nil {
		return 0
	}
	return p.CurrentStreamingSession.PlaybackPosition()
}

func (p *Player) Play() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.CurrentStreamingSession != nil {
		p.isPlaying = true
		p.CurrentStreamingSession.SetPaused(false)
		return
	}

	p.isPlaying = true
	p.play(p.CurrentTrack)
}

func (p *Player) Pause() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.CurrentStreamingSession == nil {
		return
	}

	p.CurrentStreamingSession.SetPaused(true)
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
