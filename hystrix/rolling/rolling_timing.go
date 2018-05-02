package rolling

import (
	"math"
	"sort"
	"sync"
	"time"
)

const (
	// This is the number of items we search for or sum (e.g. last 60 seconds)
	timingWindow = int64(60)

	// This is the number of items used to store data.
	// The extra item is stop collision when the window "wraps around" to the next second.
	timingItems = timingWindow + 1
)

// Timing maintains time Durations for each time bucket.
// The Durations are kept in an array to allow for a variety of
// statistics to be calculated from the source data.
type Timing struct {
	Buckets map[int64]*timingBucket
	Mutex   *sync.RWMutex

	CachedSortedDurations []time.Duration
	LastCachedTime        int64

	// allow of mocking of time in tests
	timeGeneratorSec  func() int64
	timeGeneratorNano func() int64
}

type timingBucket struct {
	timestamp int64
	Durations []time.Duration
}

// reset/empty the bucket
func (t *timingBucket) empty() {
	t.timestamp = 0
	// is there something better than this?
	t.Durations = nil
}

// NewTiming creates a RollingTiming struct.
func NewTiming() *Timing {
	r := &Timing{
		Buckets: make(map[int64]*timingBucket, timingWindow+1),
		Mutex:   &sync.RWMutex{},
	}

	// create all the buckets
	for x := int64(0); x < timingItems; x++ {
		r.Buckets[x] = &timingBucket{}
	}

	return r
}

type byDuration []time.Duration

func (c byDuration) Len() int           { return len(c) }
func (c byDuration) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c byDuration) Less(i, j int) bool { return c[i] < c[j] }

// SortedDurations returns an array of time.Duration sorted from shortest
// to longest that have occurred in the last 60 seconds.
func (r *Timing) SortedDurations() []time.Duration {
	r.Mutex.RLock()
	t := r.LastCachedTime
	r.Mutex.RUnlock()

	now := time.Now()
	nowNano := r.getTimeInNano(now)
	if t+time.Second.Nanoseconds() > nowNano {
		// don't recalculate if current cache is still fresh
		return r.CachedSortedDurations
	}

	var durations byDuration
	minTimeInSec := r.getMinTimeInSec(r.getTimeInSec(now))

	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	var b *timingBucket
	for _, b = range r.Buckets {
		if b.timestamp >= minTimeInSec {
			for _, d := range b.Durations {
				durations = append(durations, d)
			}
		}
	}

	sort.Sort(durations)

	r.CachedSortedDurations = durations
	r.LastCachedTime = nowNano

	return r.CachedSortedDurations
}

// Add appends the time.Duration given to the current time bucket.
func (r *Timing) Add(duration time.Duration) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	timeInSec := r.getTimeInSec(time.Now())
	index := r.getIndex(timeInSec)

	b := r.Buckets[index]
	if b.timestamp != timeInSec {
		// auto-empty buckets that are not clean (caused by sporadic data)
		b.empty()
		b.timestamp = timeInSec
	}

	b.Durations = append(b.Durations, duration)
}

// Percentile computes the percentile given with a linear interpolation.
func (r *Timing) Percentile(p float64) uint32 {
	sortedDurations := r.SortedDurations()
	length := len(sortedDurations)
	if length <= 0 {
		return 0
	}

	pos := r.ordinal(len(sortedDurations), p) - 1
	return uint32(sortedDurations[pos].Nanoseconds() / 1000000)
}

func (r *Timing) ordinal(length int, percentile float64) int64 {
	if percentile == 0 && length > 0 {
		return 1
	}

	return int64(math.Ceil((percentile / float64(100)) * float64(length)))
}

// Mean computes the average timing in the last 60 seconds.
func (r *Timing) Mean() uint32 {
	sortedDurations := r.SortedDurations()
	var sum time.Duration
	for _, d := range sortedDurations {
		sum += d
	}

	length := int64(len(sortedDurations))
	if length == 0 {
		return 0
	}

	return uint32(sum.Nanoseconds() / length / 1000000)
}

func (r *Timing) getTimeInSec(now time.Time) int64 {
	if r.timeGeneratorSec != nil {
		return r.timeGeneratorSec()
	}

	return now.Unix()
}

func (r *Timing) getTimeInNano(now time.Time) int64 {
	if r.timeGeneratorNano != nil {
		return r.timeGeneratorNano()
	}

	return now.UnixNano()
}

func (r *Timing) getMinTimeInSec(timeInSec int64) int64 {
	return timeInSec - timingWindow + 1
}

func (r *Timing) getIndex(timeInSec int64) int64 {
	return timeInSec % timingItems
}
