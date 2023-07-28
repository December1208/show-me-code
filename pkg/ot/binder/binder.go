package binder

import (
	"show-me-code/pkg/ot/store"
	"show-me-code/pkg/ot/text"
	"sync"
	"time"
)

type ClientMetadata struct {
	Client   interface{}
	Metadata interface{}
}

type Config struct {
	FlushPeriodMS           int64
	CloseInactivityPeriodMS int64
	OTBufferConfig          text.OTBufferConfig
}

type BinderImpl struct {
	id       string
	otBuffer text.OTBufferInterface
	store    store.Store
	config   Config

	client    []*BinderClient
	clientMux sync.Mutex
}

type BinderClient struct {
	transformChan chan<- text.Transform
	metaDataChan  chan<- ClientMetadata
}

func (b *BinderImpl) loop() {
	flushPeriod := time.Duration(b.config.FlushPeriodMS) * time.Millisecond
	closePeriod := time.Duration(b.config.CloseInactivityPeriodMS) * time.Millisecond

	flushTimer := time.NewTimer(flushPeriod)
	closeTimer := time.NewTimer(closePeriod)
	for {
		running := true
		select {
		case <-flushTimer.C:

			flushTimer.Reset(flushPeriod)
		case <-closeTimer.C:
			closeTimer.Reset(closePeriod)
		}
		if !running {

		}
	}
}
