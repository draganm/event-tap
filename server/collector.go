package server

import (
	"github.com/draganm/bolted"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
)

func newStatsCollector(db bolted.Database, log logr.Logger) prometheus.Collector {
	return &statsCollector{db: db, log: log}

}

type statsCollector struct {
	db  bolted.Database
	log logr.Logger
}

func (sc *statsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(sc, ch)
}

var (
	tapsCount = prometheus.NewDesc(
		"taps_count",
		"Number of event taps installed.",
		nil, nil,
	)
)

func (sc *statsCollector) Collect(ch chan<- prometheus.Metric) {

	var messagesCount float64

	err := bolted.SugaredRead(sc.db, func(tx bolted.SugaredReadTx) error {
		messagesCount = float64(tx.Size(tapsPath))
		return nil
	})

	if err != nil {
		sc.log.Error(err, "could not collect metrics")
	}

	ch <- prometheus.MustNewConstMetric(
		tapsCount,
		prometheus.CounterValue,
		messagesCount,
	)

}
