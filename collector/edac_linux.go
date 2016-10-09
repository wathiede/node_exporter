// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !noedac

package collector

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	edacSubsystem = "edac"
)

var (
	edacMemControllerRE = regexp.MustCompile(`.*devices/system/edac/mc/mc([0-9]*)`)
	edacMemCsrowRE      = regexp.MustCompile(`.*devices/system/edac/mc/mc[0-9]*/csrow([0-9]*)`)
)

type edacMCMetric struct {
	metricName    string
	metricType    prometheus.ValueType
	metricHelp    string
	memController string
	value         float64
}

type edacCollector struct {
	ceCount       *prometheus.Desc
	ceNoinfoCount *prometheus.Desc
	ueCount       *prometheus.Desc
	ueNoinfoCount *prometheus.Desc
	csrowCeCount  *prometheus.Desc
	csrowUeCount  *prometheus.Desc
}

func init() {
	Factories["edac"] = NewEdacCollector
}

// Takes a prometheus registry and returns a new Collector exposing
// edac stats.
func NewEdacCollector() (Collector, error) {
	return &edacCollector{
		ceCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, edacSubsystem, "correctable_errors_total"),
			"Total correctable memory errors.",
			[]string{"controller"}, nil,
		),
		ceNoinfoCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, edacSubsystem, "no_csrow_correctable_errors_total"),
			"Total correctable memory errors with no DIMM information.",
			[]string{"controller"}, nil,
		),
		ueCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, edacSubsystem, "uncorrectable_errors_total"),
			"Total uncorrectable memory errors.",
			[]string{"controller"}, nil,
		),
		ueNoinfoCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, edacSubsystem, "no_csrow_uncorrectable_errors_total"),
			"Total uncorrectable memory errors with no DIMM information.",
			[]string{"controller"}, nil,
		),
		csrowCeCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, edacSubsystem, "csrow_correctable_errors_total"),
			"Total correctable memory errors for this csrow.",
			[]string{"controller", "csrow"}, nil,
		),
		csrowUeCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, edacSubsystem, "csrow_uncorrectable_errors_total"),
			"Total correctable memory errors for this csrow.",
			[]string{"controller", "csrow"}, nil,
		),
	}, nil
}

func (c *edacCollector) Update(ch chan<- prometheus.Metric) (err error) {
	memControllers, err := filepath.Glob(sysFilePath("devices/system/edac/mc/mc[0-9]*"))
	if err != nil {
		return err
	}
	for _, controller := range memControllers {
		controllerNumber := edacMemControllerRE.FindStringSubmatch(controller)
		if controllerNumber == nil {
			return fmt.Errorf("controller number string didn't match regexp: %s", controller)
		}

		value, err := readUintFromFile(path.Join(controller, "ce_count"))
		if err != nil {
			return fmt.Errorf("couldn't get ce_count for controller %s: %s", controllerNumber, err)
		}
		ch <- prometheus.MustNewConstMetric(
			c.ceCount, prometheus.CounterValue, float64(value))

		value, err = readUintFromFile(path.Join(controller, "ce_noinfo_count"))
		if err != nil {
			return fmt.Errorf("couldn't get ce_noinfo_count for controller %s: %s", controllerNumber, err)
		}
		ch <- prometheus.MustNewConstMetric(
			c.ceNoinfoCount, prometheus.CounterValue, float64(value))

		value, err = readUintFromFile(path.Join(controller, "ue_count"))
		if err != nil {
			return fmt.Errorf("couldn't get ue_count for controller %s: %s", controllerNumber, err)
		}
		ch <- prometheus.MustNewConstMetric(
			c.ueCount, prometheus.CounterValue, float64(value))

		value, err = readUintFromFile(path.Join(controller, "ue_noinfo_count"))
		if err != nil {
			return fmt.Errorf("couldn't get ue_noinfo_count for controller %s: %s", controllerNumber, err)
		}
		ch <- prometheus.MustNewConstMetric(
			c.ueNoinfoCount, prometheus.CounterValue, float64(value))

		// For each controller, walk the csrow directories.
		csrows, err := filepath.Glob(controller + "csrow[0-9]*")
		if err != nil {
			return err
		}
		for _, csrow := range csrows {
			csrowNumber := edacMemCsrowRE.FindStringSubmatch(csrow)
			if csrowNumber == nil {
				return fmt.Errorf("controller number string didn't match regexp: %s", controller)
			}
			// TODO: Add csrow metrics
		}
	}

	return err
}
