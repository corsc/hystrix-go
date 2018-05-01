package plugins

import (
	"sync/atomic"
	"testing"

	"github.com/myteksi/hystrix-go/plugins/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestSampleRate(t *testing.T) {
	Convey("when initializing the collector", t, func() {
		Convey("with no sample rate", func() {
			client, err := InitializeStatsdCollector(&StatsdCollectorConfig{
				StatsdAddr: "localhost:8125",
				Prefix:     "test",
			})
			So(err, ShouldBeNil)

			collector := client.NewStatsdCollector("foo", "group1").(*StatsdCollector)
			Convey("it defaults to no sampling", func() {
				So(collector.sampleRate, ShouldEqual, 1.0)
			})
		})
		Convey("with a sample rate", func() {
			client, err := InitializeStatsdCollector(&StatsdCollectorConfig{
				StatsdAddr: "localhost:8125",
				Prefix:     "test",
				SampleRate: 0.5,
			})
			So(err, ShouldBeNil)

			collector := client.NewStatsdCollector("foo", "group2").(*StatsdCollector)
			Convey("the rate is set", func() {
				So(collector.sampleRate, ShouldEqual, 0.5)
			})
		})
	})
}

func TestCommandGroup(t *testing.T) {
	mockStatsd1 := &mocks.Statter{}
	mockStatsd2 := &mocks.Statter{}
	testStatsdCollector1 := &StatsdCollectorClient{}
	testStatsdCollector2 := &StatsdCollectorClient{}
	testStatsdCollector1.client = mockStatsd1
	testStatsdCollector2.client = mockStatsd2
	metricCollector1 := testStatsdCollector1.NewStatsdCollector("commandName1", "commandGroup1")
	metricCollector2 := testStatsdCollector2.NewStatsdCollector("commandName2", "commandGroup2")
	queueLength1 := int32(0)
	attempt1 := int32(0)
	queueLength2 := int32(0)

	mockStatsd1.On("Inc", "commandGroup1.commandName1.queueLength", int64(1), mock.Anything).Run(func(args mock.Arguments) {
		atomic.AddInt32(&queueLength1, 1)
	}).Return(nil)
	mockStatsd1.On("Inc", "commandGroup1.commandName1.attempts", int64(1), mock.Anything).Run(func(args mock.Arguments) {
		atomic.AddInt32(&attempt1, 1)
	}).Return(nil)
	mockStatsd2.On("Inc", "commandGroup2.commandName2.queueLength", int64(1), mock.Anything).Run(func(args mock.Arguments) {
		atomic.AddInt32(&queueLength2, 1)
	}).Return(nil)

	Convey("increment queue size", t, func() {
		metricCollector1.IncrementAttempts()
		metricCollector1.IncrementAttempts()
		metricCollector1.IncrementQueueSize()
		metricCollector2.IncrementQueueSize()

		So(2, ShouldEqual, attempt1)
		So(1, ShouldEqual, queueLength1)
		So(1, ShouldEqual, queueLength2)
	})

}
