package collector

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	Factories["stat"] = NewStatCollector
}

type statCollector struct {
	btime *prometheus.Desc
}

// NewStatCollector returns a new Collector exposing kernel/system statistics.
func NewStatCollector() (Collector, error) {
	return &statCollector{
		btime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "boot_time"),
			"Node boot time, in unixtime.",
			nil, nil,
		),
	}, nil
}

func (s *statCollector) Update(ch chan<- prometheus.Metric) error {
	type timeval struct {
		sec  int32
		usec int64
	}
	b, err := unix.SysctlRaw("kern.boottime")
	if err != nil {
		return err
	}
	bt := *(*timeval)(unsafe.Pointer((&b[0])))
	if err != nil {
		return fmt.Errorf("couldn't get boottime: %s", err)
	}
	ch <- prometheus.MustNewConstMetric(s.btime, prometheus.GaugeValue, float64(bt.sec))
	return nil
}
