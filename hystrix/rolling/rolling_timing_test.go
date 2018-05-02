package rolling

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestOrdinal(t *testing.T) {
	Convey("given a new rolling timing", t, func() {

		r := NewTiming()

		Convey("Mean() should be 0", func() {
			So(r.Mean(), ShouldEqual, 0)
		})

		Convey("and given a set of lengths and percentiles", func() {
			var ordinalTests = []struct {
				length   int
				perc     float64
				expected int64
			}{
				{1, 0, 1},
				{2, 0, 1},
				{2, 50, 1},
				{2, 51, 2},
				{5, 30, 2},
				{5, 40, 2},
				{5, 50, 3},
				{11, 25, 3},
				{11, 50, 6},
				{11, 75, 9},
				{11, 100, 11},
			}

			Convey("each should generate the expected ordinal", func() {

				for _, s := range ordinalTests {
					So(r.ordinal(s.length, s.perc), ShouldEqual, s.expected)
				}
			})
		})

		Convey("after adding 2 timings", func() {
			r.Add(100 * time.Millisecond)
			time.Sleep(2 * time.Second)
			r.Add(200 * time.Millisecond)

			Convey("the mean should be the average of the timings", func() {
				So(r.Mean(), ShouldEqual, 150)
			})
		})

		Convey("after adding many timings", func() {
			durations := []int{1, 1004, 1004, 1004, 1004, 1004, 1004, 1004, 1004, 1004, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1005, 1006, 1006, 1006, 1006, 1007, 1007, 1007, 1008, 1015}
			for _, d := range durations {
				r.Add(time.Duration(d) * time.Millisecond)
			}

			Convey("calculates correct percentiles", func() {
				So(r.Percentile(0), ShouldEqual, 1)
				So(r.Percentile(75), ShouldEqual, 1006)
				So(r.Percentile(99), ShouldEqual, 1015)
				So(r.Percentile(100), ShouldEqual, 1015)
			})
		})
	})
}
func TestTiming_2seconds(t *testing.T) {
	timing := NewTiming()

	// time generator (predictable time)
	// This one creates 2 consecutive buckets
	calls := int64(0)
	timing.timeGeneratorSec = func() int64 {
		calls++
		if calls <= 100 {
			return int64(61)
		}
		return int64(62)
	}

	timing.timeGeneratorNano = func() int64 {
		return timing.timeGeneratorSec() * time.Second.Nanoseconds()
	}

	// call add
	for x := 0; x < 150; x++ {
		timing.Add(1 * time.Second)
	}

	// validate
	assert.Equal(t, uint32(1*time.Second/1000000), timing.Percentile(1))
	assert.Equal(t, uint32(1*time.Second/1000000), timing.Percentile(25))
	assert.Equal(t, uint32(1*time.Second/1000000), timing.Percentile(50))
	assert.Equal(t, uint32(1*time.Second/1000000), timing.Percentile(75))
	assert.Equal(t, uint32(1*time.Second/1000000), timing.Percentile(99))
	assert.Equal(t, uint32(1*time.Second/1000000), timing.Mean())
}

func TestTiming_60seconds(t *testing.T) {
	timing := NewTiming()

	// time generator (predictable time)
	// This one creates 2 consecutive buckets
	calls := int64(0)
	timing.timeGeneratorSec = func() int64 {
		calls++
		if calls <= 60 {
			return 60 + calls
		}
		return 60 + 60
	}

	timing.timeGeneratorNano = func() int64 {
		return timing.timeGeneratorSec() * time.Second.Nanoseconds()
	}

	// call add
	for x := 0; x < 60; x++ {
		timing.Add(time.Duration(1+x) * time.Second)
	}

	// validate
	assert.Equal(t, uint32(1*time.Second/1000000), timing.Percentile(1))
	assert.Equal(t, uint32(15*time.Second/1000000), timing.Percentile(25))
	assert.Equal(t, uint32(30*time.Second/1000000), timing.Percentile(50))
	assert.Equal(t, uint32(45*time.Second/1000000), timing.Percentile(75))
	assert.Equal(t, uint32(60*time.Second/1000000), timing.Percentile(99))
	// mean of 1,2,...,60
	assert.Equal(t, uint32(float64(30.5)*float64(time.Second)/1000000), timing.Mean())
}

func TestTiming_100seconds(t *testing.T) {
	timing := NewTiming()

	// time generator (predictable time)
	// This one creates 2 consecutive buckets
	calls := int64(0)
	timing.timeGeneratorSec = func() int64 {
		calls++
		if calls <= 100 {
			return 60 + calls
		}
		return 60 + 100
	}

	timing.timeGeneratorNano = func() int64 {
		return timing.timeGeneratorSec() * time.Second.Nanoseconds()
	}

	// call add
	for x := 0; x < 100; x++ {
		timing.Add(time.Duration(1+x) * time.Second)
	}

	// validate
	assert.Equal(t, uint32(41*time.Second/1000000), timing.Percentile(1), fmt.Sprintf("expected %.3f; was %.3f", float64(41*time.Second/1000000), float64(timing.Percentile(1))))
	assert.Equal(t, uint32(55*time.Second/1000000), timing.Percentile(25))
	assert.Equal(t, uint32(70*time.Second/1000000), timing.Percentile(50))
	assert.Equal(t, uint32(85*time.Second/1000000), timing.Percentile(75))
	assert.Equal(t, uint32(100*time.Second/1000000), timing.Percentile(99))
	// mean of 41,42,...,100
	assert.Equal(t, uint32(float64(70.5)*float64(time.Second)/1000000), timing.Mean(), fmt.Sprintf("expected %.3f; was %.3f", float64(70.5)*float64(time.Second/1000000), float64(timing.Mean())))
}
