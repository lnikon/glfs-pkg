package server

// import (
// 	"fmt"
// 	"time"

// 	metrics "github.com/go-kit/kit/metrics"
// )

// type instrumentingMiddleware struct {
// 	requestCount   metrics.Counter
// 	requestLatency metrics.Histogram
// 	countResult    metrics.Histogram
// }

// func (mw instrumentingMiddleware) Uppercase(s string) (output string, err error) {
// 	defer func(begin time.Time) {
// 		lvs := []string{"method", "uppercase", "error", fmt.Sprint(err != nil)}
// 		mw.requestCount.With(lvs...).Add(1)
// 		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
// 	}(time.Now())

// 	output, err = mw.next.Uppercase(s)
// 	return
// }

// func (mw instrumentingMiddleware) Count(s string) (output int) {
// 	defer func(begin time.Time) {
// 		lvs := []string{"method", "count", "error"}
// 		mw.requestCount.With(lvs...).Add(1)
// 		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
// 	}(time.Now())

// 	output = mw.next.Count(s)
// 	return
// }
