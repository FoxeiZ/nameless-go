package player

import (
	"sync"

	"github.com/foxeiz/namelessgo/src/extractors"
)

type Queue struct {
	sync.Mutex
	TrackList []*extractors.TrackInfo
}

func NewQueue() *Queue {
	return &Queue{
		TrackList: make([]*extractors.TrackInfo, 0),
	}
}

func (q *Queue) Enqueue(t *extractors.TrackInfo) {
	q.Lock()
	defer q.Unlock()
	q.TrackList = append(q.TrackList, t)
}

func (q *Queue) Dequeue() *extractors.TrackInfo {
	q.Lock()
	defer q.Unlock()
	if len(q.TrackList) > 0 {
		x := q.TrackList[0]
		q.TrackList = q.TrackList[1:]
		return x
	}
	return nil
}

func (q *Queue) Pop(x int) *extractors.TrackInfo {
	q.Lock()
	defer q.Unlock()
	if len(q.TrackList) > 0 {
		track := q.TrackList[x]
		q.TrackList = append(q.TrackList[:x], q.TrackList[x+1:]...)
		return track
	}
	return nil
}

func (q *Queue) Peek(x int) *extractors.TrackInfo {
	q.Lock()
	defer q.Unlock()
	if len(q.TrackList) > 0 && x < len(q.TrackList) {
		track := q.TrackList[x]
		return track
	}
	return nil
}

func (q *Queue) GetTrackList() []*extractors.TrackInfo {
	q.Lock()
	defer q.Unlock()
	return q.TrackList
}

func (q *Queue) GetTrackListSlice(x int, y int) []*extractors.TrackInfo {
	q.Lock()
	defer q.Unlock()

	if y > len(q.TrackList) {
		y = len(q.TrackList)
	}
	if x > len(q.TrackList) {
		x = len(q.TrackList)
	}

	slice := q.TrackList[x:y]
	return slice
}

func (q *Queue) Length() int {
	q.Lock()
	defer q.Unlock()
	return len(q.TrackList)
}

func (q *Queue) IsEmpty() bool {
	q.Lock()
	defer q.Unlock()
	return len(q.TrackList) == 0
}

func (q *Queue) Clear() {
	q.Lock()
	defer q.Unlock()
	q.TrackList = make([]*extractors.TrackInfo, 0)
}
