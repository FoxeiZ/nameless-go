package music_commands

import (
	"sync"
)

type Queue struct {
	Mutex     *sync.Mutex
	TrackList []*TrackInfo
}

func NewQueue() *Queue {
	return &Queue{&sync.Mutex{}, make([]*TrackInfo, 0)}
}

func (q *Queue) Enqueue(t *TrackInfo) {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	q.TrackList = append(q.TrackList, t)
}

func (q *Queue) Dequeue() *TrackInfo {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	if len(q.TrackList) > 0 {
		x := q.TrackList[0]
		q.TrackList = q.TrackList[1:]
		return x
	}
	return nil
}

func (q *Queue) Pop(x int) *TrackInfo {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	if len(q.TrackList) > 0 {
		track := q.TrackList[x]
		q.TrackList = append(q.TrackList[:x], q.TrackList[x+1:]...)
		return track
	}
	return nil
}

func (q *Queue) Peek(x int) *TrackInfo {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	if len(q.TrackList) > 0 && x < len(q.TrackList) {
		track := q.TrackList[x]
		q.Mutex.Unlock()
		return track
	}
	return nil
}

func (q *Queue) GetTrackList() []*TrackInfo {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	return q.TrackList
}

func (q *Queue) GetTrackListSlice(x int, y int) []*TrackInfo {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()

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
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	return len(q.TrackList)
}

func (q *Queue) IsEmpty() bool {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	return len(q.TrackList) == 0
}

func (q *Queue) Clear() {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	q.TrackList = make([]*TrackInfo, 0)
}
