package rolling

import (
	"sync"
	"time"
)

const (
	// This is the number of items we search for or sum (e.g. last 10 seconds)
	numberWindow = int64(10)

	// This is the number of items used to store data.
	// The extra item is stop collision when the window "wraps around" to the next second.
	numberItems = numberWindow + 1
)

// Number tracks a numberBucket over a bounded number of
// time buckets. Currently the buckets are one second long and only the last 10 seconds are kept.
type Number struct {
	Buckets map[int64]*numberBucket
	Mutex   *sync.RWMutex

	// allow of mocking of time in tests
	timeGenerator func() int64
}

type numberBucket struct {
	timestamp int64
	Value     float64
}

// reset/empty the bucket
func (n *numberBucket) empty() {
	n.timestamp = 0
	n.Value = 0
}

// NewNumber initializes a RollingNumber struct.
func NewNumber() *Number {
	r := &Number{
		// keep only 60 seconds worth of buckets and never recreate them
		Buckets: make(map[int64]*numberBucket, numberItems),
		Mutex:   &sync.RWMutex{},
	}

	// create all the buckets
	for x := int64(0); x < numberItems; x++ {
		r.Buckets[x] = &numberBucket{}
	}

	return r
}

// Increment increments the number in current timeBucket.
func (r *Number) Increment(i float64) {
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

	b.Value += i
}

// UpdateMax updates the maximum value in the current bucket.
func (r *Number) UpdateMax(n float64) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	timeInSec := r.getTimeInSec(time.Now())
	index := r.getIndex(timeInSec)

	b := r.Buckets[index]
	if b.timestamp != timeInSec {
		b.empty()
		b.timestamp = timeInSec
	}

	// only use those buckets that are within the time box; we cannot empty them without a write lock
	if n > b.Value {
		b.Value = n
	}
}

// Sum sums the values over the buckets in the last 10 seconds.
func (r *Number) Sum(in time.Time) float64 {
	sum := float64(0)

	r.Mutex.RLock()
	defer r.Mutex.RUnlock()

	minTimeInSec := r.getMinTimeInSec(r.getTimeInSec(in))

	// to sum the "window" we sum all except the next (extra) one
	for _, b := range r.Buckets {
		if b.timestamp >= minTimeInSec {
			// only use those buckets that are within the time box; we cannot empty them without a write lock
			sum += b.Value
		}
	}

	return sum
}

// Max returns the maximum value seen in the last 10 seconds.
func (r *Number) Max(in time.Time) float64 {
	var max float64

	r.Mutex.RLock()
	defer r.Mutex.RUnlock()

	timeInSec := r.getTimeInSec(in)
	minTimeInSec := r.getMinTimeInSec(timeInSec)

	var b *numberBucket
	for _, b = range r.Buckets {
		if b.timestamp >= minTimeInSec {
			// only use those buckets that are within the time box; we cannot empty them without a write lock
			if b.Value > max {
				max = b.Value
			}
		}
	}

	return max
}

// Avg return the average value seen in the last 10 seconds.
func (r *Number) Avg(in time.Time) float64 {
	return r.Sum(in) / float64(numberWindow)
}

func (r *Number) getTimeInSec(now time.Time) int64 {
	if r.timeGenerator != nil {
		return r.timeGenerator()
	}

	return now.Unix()
}

func (r *Number) getMinTimeInSec(timeInSec int64) int64 {
	return timeInSec - numberWindow + 1
}

func (r *Number) getIndex(timeInSec int64) int64 {
	return timeInSec % numberItems
}
