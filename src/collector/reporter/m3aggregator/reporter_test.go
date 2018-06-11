// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package m3aggregator

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/m3db/m3aggregator/client"
	"github.com/m3db/m3metrics/aggregation"
	"github.com/m3db/m3metrics/matcher"
	"github.com/m3db/m3metrics/metadata"
	"github.com/m3db/m3metrics/metric/id"
	"github.com/m3db/m3metrics/metric/unaggregated"
	"github.com/m3db/m3metrics/policy"
	"github.com/m3db/m3metrics/rules"
	"github.com/m3db/m3x/clock"
	"github.com/m3db/m3x/instrument"
	xtime "github.com/m3db/m3x/time"

	"github.com/fortytw2/leaktest"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
)

var (
	testMatchResult = rules.NewMatchResult(
		0,
		math.MaxInt64,
		metadata.StagedMetadatas{
			{
				CutoverNanos: 0,
				Tombstoned:   false,
				Metadata: metadata.Metadata{
					Pipelines: []metadata.PipelineMetadata{
						{
							AggregationID: aggregation.DefaultID,
							StoragePolicies: policy.StoragePolicies{
								policy.NewStoragePolicy(20*time.Second, xtime.Second, 6*time.Hour),
								policy.NewStoragePolicy(10*time.Minute, xtime.Minute, 25*24*time.Hour),
							},
						},
						{
							AggregationID: aggregation.MustCompressTypes(aggregation.Max),
							StoragePolicies: policy.StoragePolicies{
								policy.NewStoragePolicy(time.Minute, xtime.Minute, 2*24*time.Hour),
							},
						},
					},
				},
			},
			{
				CutoverNanos: math.MaxInt64,
				Tombstoned:   false,
				Metadata: metadata.Metadata{
					Pipelines: []metadata.PipelineMetadata{
						{
							AggregationID: aggregation.MustCompressTypes(aggregation.Max, aggregation.P9999),
							StoragePolicies: policy.StoragePolicies{
								policy.NewStoragePolicy(time.Second, xtime.Second, time.Hour),
							},
						},
					},
				},
			},
		},
		[]rules.IDWithMetadatas{
			{
				ID:        []byte("foo"),
				Metadatas: metadata.DefaultStagedMetadatas,
			},
			{
				ID: []byte("bar"),
				Metadatas: metadata.StagedMetadatas{
					{
						CutoverNanos: 100,
						Tombstoned:   false,
						Metadata: metadata.Metadata{
							Pipelines: []metadata.PipelineMetadata{
								{
									AggregationID: aggregation.DefaultID,
									StoragePolicies: policy.StoragePolicies{
										policy.NewStoragePolicy(time.Minute, xtime.Minute, 2*24*time.Hour),
										policy.NewStoragePolicy(10*time.Minute, xtime.Minute, 25*24*time.Hour),
									},
								},
								{
									AggregationID: aggregation.MustCompressTypes(aggregation.Max, aggregation.P9999),
									StoragePolicies: policy.StoragePolicies{
										policy.NewStoragePolicy(20*time.Second, xtime.Second, 6*time.Hour),
									},
								},
							},
						},
					},
					{
						CutoverNanos: 200,
						Tombstoned:   true,
						Metadata: metadata.Metadata{
							Pipelines: []metadata.PipelineMetadata{
								{
									AggregationID: aggregation.MustCompressTypes(aggregation.P9999),
									StoragePolicies: policy.StoragePolicies{
										policy.NewStoragePolicy(time.Second, xtime.Second, time.Hour),
									},
								},
							},
						},
					},
				},
			},
		})
)

func TestReporterReportCounter(t *testing.T) {
	defer leaktest.Check(t)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		now              = time.Unix(1234, 0)
		nowFn            = func() time.Time { return now }
		positiveSkew     = 10 * time.Second
		negativeSkew     = 10 * time.Second
		fromNanos        = now.Add(-negativeSkew).UnixNano()
		toNanos          = now.Add(positiveSkew).UnixNano()
		errReportCounter = errors.New("test report counter error")
		actual           []unaggregated.CounterWithMetadatas
	)
	clockOpts := clock.NewOptions().
		SetNowFn(nowFn).
		SetMaxPositiveSkew(positiveSkew).
		SetMaxNegativeSkew(negativeSkew)
	mockID := id.NewMockID(ctrl)
	mockID.EXPECT().Bytes().Return([]byte("testCounter"))
	mockMatcher := matcher.NewMockMatcher(ctrl)
	mockMatcher.EXPECT().ForwardMatch(mockID, fromNanos, toNanos).Return(testMatchResult)
	mockMatcher.EXPECT().Close().Return(nil).AnyTimes()
	mockClient := client.NewMockClient(ctrl)
	mockClient.EXPECT().
		WriteUntimedCounter(gomock.Any(), gomock.Any()).
		DoAndReturn(func(counter unaggregated.Counter, metadatas metadata.StagedMetadatas) error {
			actual = append(actual, unaggregated.CounterWithMetadatas{
				Counter:         counter,
				StagedMetadatas: metadatas,
			})
			return errReportCounter
		}).MinTimes(1)
	mockClient.EXPECT().Close().Return(nil).AnyTimes()
	reporter := NewReporter(mockMatcher, mockClient, NewReporterOptions().SetClockOptions(clockOpts))
	defer reporter.Close()
	err := reporter.ReportCounter(mockID, 1234)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), errReportCounter.Error()))

	expected := []unaggregated.CounterWithMetadatas{
		{
			Counter: unaggregated.Counter{
				ID:    []byte("testCounter"),
				Value: 1234,
			},
			StagedMetadatas: metadata.StagedMetadatas{
				{
					CutoverNanos: 0,
					Tombstoned:   false,
					Metadata: metadata.Metadata{
						Pipelines: []metadata.PipelineMetadata{
							{
								AggregationID: aggregation.DefaultID,
								StoragePolicies: policy.StoragePolicies{
									policy.NewStoragePolicy(20*time.Second, xtime.Second, 6*time.Hour),
									policy.NewStoragePolicy(10*time.Minute, xtime.Minute, 25*24*time.Hour),
								},
							},
							{
								AggregationID: aggregation.MustCompressTypes(aggregation.Max),
								StoragePolicies: policy.StoragePolicies{
									policy.NewStoragePolicy(time.Minute, xtime.Minute, 2*24*time.Hour),
								},
							},
						},
					},
				},
				{
					CutoverNanos: math.MaxInt64,
					Tombstoned:   false,
					Metadata: metadata.Metadata{
						Pipelines: []metadata.PipelineMetadata{
							{
								AggregationID: aggregation.MustCompressTypes(aggregation.Max, aggregation.P9999),
								StoragePolicies: policy.StoragePolicies{
									policy.NewStoragePolicy(time.Second, xtime.Second, time.Hour),
								},
							},
						},
					},
				},
			},
		},
		{
			Counter: unaggregated.Counter{
				ID:    []byte("foo"),
				Value: 1234,
			},
			StagedMetadatas: metadata.DefaultStagedMetadatas,
		},
	}
	require.Equal(t, expected, actual)
}

func TestReporterReportBatchTimer(t *testing.T) {
	defer leaktest.Check(t)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		now                 = time.Unix(1234, 0)
		nowFn               = func() time.Time { return now }
		positiveSkew        = 10 * time.Second
		negativeSkew        = 10 * time.Second
		fromNanos           = now.Add(-negativeSkew).UnixNano()
		toNanos             = now.Add(positiveSkew).UnixNano()
		errReportBatchTimer = errors.New("test report batch timer error")
		actual              []unaggregated.BatchTimerWithMetadatas
	)
	clockOpts := clock.NewOptions().
		SetNowFn(nowFn).
		SetMaxPositiveSkew(positiveSkew).
		SetMaxNegativeSkew(negativeSkew)
	mockID := id.NewMockID(ctrl)
	mockID.EXPECT().Bytes().Return([]byte("testBatchTimer"))
	mockMatcher := matcher.NewMockMatcher(ctrl)
	mockMatcher.EXPECT().ForwardMatch(mockID, fromNanos, toNanos).Return(testMatchResult)
	mockMatcher.EXPECT().Close().Return(nil).AnyTimes()
	mockClient := client.NewMockClient(ctrl)
	mockClient.EXPECT().
		WriteUntimedBatchTimer(gomock.Any(), gomock.Any()).
		DoAndReturn(
			func(
				batchTimer unaggregated.BatchTimer,
				metadatas metadata.StagedMetadatas,
			) error {
				actual = append(actual, unaggregated.BatchTimerWithMetadatas{
					BatchTimer:      batchTimer,
					StagedMetadatas: metadatas,
				})
				return errReportBatchTimer
			}).
		MinTimes(1)
	mockClient.EXPECT().Close().Return(nil).AnyTimes()
	reporter := NewReporter(mockMatcher, mockClient, NewReporterOptions().SetClockOptions(clockOpts))
	defer reporter.Close()
	err := reporter.ReportBatchTimer(mockID, []float64{1.3, 2.4})
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), errReportBatchTimer.Error()))

	expected := []unaggregated.BatchTimerWithMetadatas{
		{
			BatchTimer: unaggregated.BatchTimer{
				ID:     []byte("testBatchTimer"),
				Values: []float64{1.3, 2.4},
			},
			StagedMetadatas: metadata.StagedMetadatas{
				{
					CutoverNanos: 0,
					Tombstoned:   false,
					Metadata: metadata.Metadata{
						Pipelines: []metadata.PipelineMetadata{
							{
								AggregationID: aggregation.DefaultID,
								StoragePolicies: policy.StoragePolicies{
									policy.NewStoragePolicy(20*time.Second, xtime.Second, 6*time.Hour),
									policy.NewStoragePolicy(10*time.Minute, xtime.Minute, 25*24*time.Hour),
								},
							},
							{
								AggregationID: aggregation.MustCompressTypes(aggregation.Max),
								StoragePolicies: policy.StoragePolicies{
									policy.NewStoragePolicy(time.Minute, xtime.Minute, 2*24*time.Hour),
								},
							},
						},
					},
				},
				{
					CutoverNanos: math.MaxInt64,
					Tombstoned:   false,
					Metadata: metadata.Metadata{
						Pipelines: []metadata.PipelineMetadata{
							{
								AggregationID: aggregation.MustCompressTypes(aggregation.Max, aggregation.P9999),
								StoragePolicies: policy.StoragePolicies{
									policy.NewStoragePolicy(time.Second, xtime.Second, time.Hour),
								},
							},
						},
					},
				},
			},
		},
		{
			BatchTimer: unaggregated.BatchTimer{
				ID:     []byte("foo"),
				Values: []float64{1.3, 2.4},
			},
			StagedMetadatas: metadata.DefaultStagedMetadatas,
		},
	}
	require.Equal(t, expected, actual)
}

func TestReporterReportGauge(t *testing.T) {
	defer leaktest.Check(t)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		now              = time.Unix(1234, 0)
		nowFn            = func() time.Time { return now }
		positiveSkew     = 10 * time.Second
		negativeSkew     = 10 * time.Second
		fromNanos        = now.Add(-negativeSkew).UnixNano()
		toNanos          = now.Add(positiveSkew).UnixNano()
		errReportCounter = errors.New("test report gauge error")
		actual           []unaggregated.GaugeWithMetadatas
	)
	clockOpts := clock.NewOptions().
		SetNowFn(nowFn).
		SetMaxPositiveSkew(positiveSkew).
		SetMaxNegativeSkew(negativeSkew)
	mockID := id.NewMockID(ctrl)
	mockID.EXPECT().Bytes().Return([]byte("testCounter"))
	mockMatcher := matcher.NewMockMatcher(ctrl)
	mockMatcher.EXPECT().ForwardMatch(mockID, fromNanos, toNanos).Return(testMatchResult)
	mockMatcher.EXPECT().Close().Return(nil).AnyTimes()
	mockClient := client.NewMockClient(ctrl)
	mockClient.EXPECT().
		WriteUntimedGauge(gomock.Any(), gomock.Any()).
		DoAndReturn(func(gauge unaggregated.Gauge, metadatas metadata.StagedMetadatas) error {
			actual = append(actual, unaggregated.GaugeWithMetadatas{
				Gauge:           gauge,
				StagedMetadatas: metadatas,
			})
			return errReportCounter
		}).MinTimes(1)
	mockClient.EXPECT().Close().Return(nil).AnyTimes()
	reporter := NewReporter(mockMatcher, mockClient, NewReporterOptions().SetClockOptions(clockOpts))
	defer reporter.Close()
	err := reporter.ReportGauge(mockID, 1.8)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), errReportCounter.Error()))

	expected := []unaggregated.GaugeWithMetadatas{
		{
			Gauge: unaggregated.Gauge{
				ID:    []byte("testCounter"),
				Value: 1.8,
			},
			StagedMetadatas: metadata.StagedMetadatas{
				{
					CutoverNanos: 0,
					Tombstoned:   false,
					Metadata: metadata.Metadata{
						Pipelines: []metadata.PipelineMetadata{
							{
								AggregationID: aggregation.DefaultID,
								StoragePolicies: policy.StoragePolicies{
									policy.NewStoragePolicy(20*time.Second, xtime.Second, 6*time.Hour),
									policy.NewStoragePolicy(10*time.Minute, xtime.Minute, 25*24*time.Hour),
								},
							},
							{
								AggregationID: aggregation.MustCompressTypes(aggregation.Max),
								StoragePolicies: policy.StoragePolicies{
									policy.NewStoragePolicy(time.Minute, xtime.Minute, 2*24*time.Hour),
								},
							},
						},
					},
				},
				{
					CutoverNanos: math.MaxInt64,
					Tombstoned:   false,
					Metadata: metadata.Metadata{
						Pipelines: []metadata.PipelineMetadata{
							{
								AggregationID: aggregation.MustCompressTypes(aggregation.Max, aggregation.P9999),
								StoragePolicies: policy.StoragePolicies{
									policy.NewStoragePolicy(time.Second, xtime.Second, time.Hour),
								},
							},
						},
					},
				},
			},
		},
		{
			Gauge: unaggregated.Gauge{
				ID:    []byte("foo"),
				Value: 1.8,
			},
			StagedMetadatas: metadata.DefaultStagedMetadatas,
		},
	}
	require.Equal(t, expected, actual)
}

func TestReporterFlush(t *testing.T) {
	defer leaktest.Check(t)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var numFlushes int
	mockMatcher := matcher.NewMockMatcher(ctrl)
	mockMatcher.EXPECT().Close().Return(nil).AnyTimes()
	mockClient := client.NewMockClient(ctrl)
	mockClient.EXPECT().
		Flush().
		DoAndReturn(func() error {
			numFlushes++
			return nil
		})
	mockClient.EXPECT().Close().Return(nil).AnyTimes()
	reporter := NewReporter(mockMatcher, mockClient, NewReporterOptions())
	defer reporter.Close()

	require.NoError(t, reporter.Flush())
	require.Equal(t, 1, numFlushes)
}

func TestReporterClose(t *testing.T) {
	defer leaktest.Check(t)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	errClientClose := errors.New("test client close error")
	mockMatcher := matcher.NewMockMatcher(ctrl)
	mockMatcher.EXPECT().Close().Return(nil)
	mockClient := client.NewMockClient(ctrl)
	mockClient.EXPECT().Close().Return(errClientClose)
	reporter := NewReporter(mockMatcher, mockClient, NewReporterOptions())
	err := reporter.Close()
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), errClientClose.Error()))
}

func TestReporterMultipleCloses(t *testing.T) {
	defer leaktest.Check(t)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	numWorkers := 10
	mockMatcher := matcher.NewMockMatcher(ctrl)
	mockMatcher.EXPECT().Close().Return(nil)
	mockClient := client.NewMockClient(ctrl)
	mockClient.EXPECT().Close().Return(nil)
	r := NewReporter(mockMatcher, mockClient, NewReporterOptions())

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Close()
		}()
	}
	wg.Wait()
	require.Equal(t, int32(1), r.(*reporter).closed)
}

func TestReporterReportPending(t *testing.T) {
	defer leaktest.Check(t)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMatcher := matcher.NewMockMatcher(ctrl)
	mockMatcher.EXPECT().Close().Return(nil)
	mockClient := client.NewMockClient(ctrl)
	mockClient.EXPECT().Close().Return(nil)

	reportInterval := 50 * time.Millisecond
	scope := tally.NewTestScope("", map[string]string{"component": "reporter"})
	instrumentOpts := instrument.NewOptions().
		SetMetricsScope(scope).
		SetReportInterval(reportInterval)
	opts := NewReporterOptions().SetInstrumentOptions(instrumentOpts)
	r := NewReporter(mockMatcher, mockClient, opts).(*reporter)
	defer r.Close()
	require.Equal(t, int64(0), r.currentReportPending())

	hostname, err := os.Hostname()
	require.NoError(t, err)
	expectedID := fmt.Sprintf("report-pending+component=reporter,host=%s", hostname)

	// Increment report pending and wait for the metric to be reported.
	iter := 10
	for i := 0; i < iter; i++ {
		r.incrementReportPending()
	}
	require.Equal(t, int64(iter), r.currentReportPending())
	time.Sleep(2 * reportInterval)
	gauges := scope.Snapshot().Gauges()
	require.Equal(t, 1, len(gauges))
	res, exists := gauges[expectedID]
	require.True(t, exists)
	require.Equal(t, float64(iter), res.Value())

	// Decrement report pending and wait for the metric to be reported.
	for i := 0; i < iter; i++ {
		r.decrementReportPending()
	}
	require.Equal(t, int64(0), r.currentReportPending())
	time.Sleep(2 * reportInterval)
	gauges = scope.Snapshot().Gauges()
	require.Equal(t, 1, len(gauges))
	res, exists = gauges[expectedID]
	require.True(t, exists)
	require.Equal(t, 0.0, res.Value())
}
