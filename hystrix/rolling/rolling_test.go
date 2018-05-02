package rolling

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"strconv"
)

func TestMax(t *testing.T) {
	Convey("when adding values to a rolling number", t, func() {
		n := NewNumber()
		for _, x := range []float64{10, 11, 9} {
			n.UpdateMax(x)
			time.Sleep(1 * time.Second)
		}

		Convey("it should know the maximum", func() {
			So(n.Max(time.Now()), ShouldEqual, 11)
		})
	})
}

func TestAvg(t *testing.T) {
	Convey("when adding values to a rolling number", t, func() {
		n := NewNumber()
		for _, x := range []float64{0.5, 1.5, 2.5, 3.5, 4.5} {
			n.Increment(x)
			time.Sleep(1 * time.Second)
		}

		Convey("it should calculate the average over the number of configured buckets", func() {
			So(n.Avg(time.Now()), ShouldEqual, 1.25)
		})
	})
}

func BenchmarkRollingNumber(b *testing.B) {
	n := NewNumber()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n.Increment(1)
		_ = n.Avg(time.Now())
	}
}

func BenchmarkRollingNumberUpdateMax(b *testing.B) {
	n := NewNumber()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n.UpdateMax(float64(i))
	}
}

func TestNumber_2seconds(t *testing.T) {
	number := NewNumber()

	// time generator (predictable time)
	calls := int64(0)
	number.timeGenerator = func() int64 {
		calls++
		if calls <= 100 {
			return 60 + 1
		}
		return 60 + 2
	}

	// call increment a lot
	for x := 0; x < 150; x++ {
		number.Increment(1)
	}

	// validate
	assert.Equal(t, float64(100+50), number.Sum(time.Time{}))
	assert.Equal(t, float64(150/10), number.Avg(time.Time{}))
	assert.Equal(t, float64(100), number.Max(time.Time{}))
}

func TestNumber_100seconds(t *testing.T) {
	number := NewNumber()

	// time generator (predictable time)
	calls := int64(0)
	number.timeGenerator = func() int64 {
		calls++
		if calls <= 100 {
			return 60 + calls
		}
		return 60 + 100
	}

	// call increment a lot
	for x := 0; x < 100; x++ {
		number.Increment(1)
	}

	// validate
	assert.Equal(t, float64(10), number.Sum(time.Time{}))
	assert.Equal(t, float64(10/10), number.Avg(time.Time{}))
	assert.Equal(t, float64(1), number.Max(time.Time{}))
}

func TestNumber_10seconds(t *testing.T) {
	number := NewNumber()

	// time generator (predictable time)
	calls := int64(0)
	number.timeGenerator = func() int64 {
		calls++
		if calls <= 10 {
			return 60 + calls
		}
		return 60 + 10
	}

	// call increment a lot
	for x := 0; x < 10; x++ {
		number.Increment(float64((x + 1) * 10))
	}

	// validate
	assert.Equal(t, float64(550), number.Sum(time.Time{}))
	assert.Equal(t, float64(550/10), number.Avg(time.Time{}))
	assert.Equal(t, float64(100), number.Max(time.Time{}))
}

func TestNumber_getIndex(t *testing.T) {
	scenarios := []struct {
		in       int64
		expected int64
	}{
		{
			in:       60,
			expected: 5,
		},
		{
			in:       61,
			expected: 6,
		},
		{
			in:       62,
			expected: 7,
		},
		{
			in:       63,
			expected: 8,
		},
		{
			in:       64,
			expected: 9,
		},
		{
			in:       65,
			expected: 10,
		},
		{
			in:       66,
			expected: 0,
		},
		{
			in:       67,
			expected: 1,
		},
	}

	for _, scenario := range scenarios {
		t.Run(strconv.FormatInt(scenario.in, 10), func(t *testing.T) {
			n := NewNumber()
			result := n.getIndex(scenario.in)
			assert.Equal(t, scenario.expected, result, strconv.FormatInt(scenario.in, 10))
		})
	}

}

func TestNumber_getTimeInSec(t *testing.T) {
	scenarios := []struct {
		in       time.Time
		expected int64
	}{
		{
			in:       time.Date(2000, time.January, 03, 10, 11, 12, 0, time.UTC),
			expected: 946894272,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.in.Format(time.RFC3339), func(t *testing.T) {
			n := NewNumber()
			result := n.getTimeInSec(scenario.in)
			assert.Equal(t, scenario.expected, result, scenario.in.Format(time.RFC3339))
		})
	}

}
func TestNumber_getMinTimeInSec(t *testing.T) {
	scenarios := []struct {
		in       int64
		expected int64
	}{
		{
			in:       time.Date(2000, time.January, 03, 10, 11, 12, 0, time.UTC).Unix(),
			expected: time.Date(2000, time.January, 03, 10, 11, 3, 0, time.UTC).Unix(),
		},
	}

	for _, scenario := range scenarios {
		t.Run(strconv.FormatInt(scenario.in, 10), func(t *testing.T) {
			n := NewNumber()
			result := n.getMinTimeInSec(scenario.in)
			assert.Equal(t, scenario.expected, result, strconv.FormatInt(scenario.in, 10))
		})
	}

}
