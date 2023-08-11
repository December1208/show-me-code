package binder

import (
	"show-me-code/pkg/ot/store"
	"show-me-code/pkg/ot/text"
	"sync"
	"time"
)

type subscribeRequest struct {
	// metadata   interface{}

	portalChan chan<- *PortalImpl
	errChan    chan<- error
}

type Config struct {
	FlushPeriodMS           int64
	CloseInactivityPeriodMS int64
	OTBufferConfig          text.OTBufferConfig
}

type BinderImpl struct {
	id       string
	otBuffer text.Type
	store    store.Type
	config   Config

	client        []*binderClient
	subscribeChan chan subscribeRequest
	clientMux     sync.Mutex

	// Control Channels
	closedChan chan struct{}
	errorChan  chan<- Error
}

func NewBinder(
	id string,
	store store.Type,
	config Config,
	errorChan chan<- Error,
) (Type, error) {

	impl := BinderImpl{
		id:            id,
		store:         store,
		config:        config,
		client:        make([]*binderClient, 0),
		subscribeChan: make(chan subscribeRequest),
		closedChan:    make(chan struct{}),
		errorChan:     errorChan,
	}

	doc, err := store.Read(id)
	if err != nil {
		return nil, err
	}

	impl.otBuffer = text.NewOTBuffer(doc.Content, impl.config.OTBufferConfig)
	go impl.loop()

	return &impl, nil
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

func (b *BinderImpl) Close() {
	close(b.subscribeChan)
	<-b.closedChan
}
