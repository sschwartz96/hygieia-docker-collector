package internal

import "go.uber.org/zap"

type Collector struct {
	logger *zap.Logger
	store  Store
}

func NewCollector(logger *zap.Logger, store Store) *Collector {
	return &Collector{logger: logger, store: store}
}

func (c *Collector) Collect() error {
	// TODO: implement
	return nil
}
