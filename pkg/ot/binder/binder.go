package binder

import (
	"errors"
	"show-me-code/pkg/ot/store"
	"show-me-code/pkg/ot/text"
	"show-me-code/pkg/util"
	"sync"
	"time"
)

var (
	TimeoutError = errors.New("subscribe timeout error.")
)

type subscribeRequest struct {
	// metadata   interface{}

	portalChan chan<- *PortalImpl
	errChan    chan<- error
}

type Config struct {
	FlushPeriodMS           int64               // 刷新间隔
	CloseInactivityPeriodMS int64               // 关闭间隔
	SubscribeTimeoutMS      int64               // 订阅超时时间
	OTBufferConfig          text.OTBufferConfig // buffer 配置
}

// BinderImpl, 关联 doc 和 client
type BinderImpl struct {
	id       string     // id
	otBuffer text.Type  // buffer struct
	store    store.Type // store struct
	config   Config     // 配置

	clients       []*binderClient          // 客户端
	subscribeChan chan subscribeRequest    // 订阅请求channel
	clientMux     sync.Mutex               // 客户端锁
	transformChan chan transformSubmission // 请求channel

	// Control Channels
	closedChan chan struct{}      // 关闭 channel
	errorChan  chan<- Error       // 异常 channel
	exitChan   chan *binderClient // client 退出channel
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
		clients:       make([]*binderClient, 0),
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
		case req, open := <-b.subscribeChan:
			if open {
				b.processSubscriber(req)
			} else {
				running = false
			}
		case req, open := <-b.transformChan:
			if open {
				b.processTransform(req)
			} else {
				running = false
			}
		}

		if !running {

		}
	}
}

func (b *BinderImpl) Subscribe(timeout time.Duration) (Portal, error) {
	portalChan, errChan := make(chan *PortalImpl, 1), make(chan error, 1)
	req := subscribeRequest{
		portalChan: portalChan,
		errChan:    errChan,
	}

	select {
	case b.subscribeChan <- req:
	case <-time.After(timeout):
		return nil, TimeoutError
	}
	select {
	case portal := <-portalChan:
		return portal, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(timeout):
	}
	return nil, TimeoutError
}

func (b *BinderImpl) processSubscriber(req subscribeRequest) error {
	transformSendChan := make(chan text.Transform, 1)
	_, err := b.flush()
	if err != nil {
		select {
		case req.errChan <- err:
		default:
		}
		return err
	}

	client := binderClient{
		metadata:          nil,
		transformSendChan: transformSendChan,
	}
	portal := PortalImpl{
		client:            &client,
		transformRcvChan:  transformSendChan,
		transformSendChan: b.transformChan,
		exitChan:          b.exitChan,
	}
	select {
	case req.portalChan <- &portal:
		b.clients = append(b.clients, &client)
	case <-time.After(time.Duration(b.config.SubscribeTimeoutMS)):
		util.Logger.Info("subscribe err.")
	}
	return nil
}

func (b *BinderImpl) processTransform(req transformSubmission) {

}

func (b *BinderImpl) flush() (store.Document, error) {

	doc, err := b.store.Read(b.id)
	return doc, err
}

func (b *BinderImpl) Close() {
	close(b.subscribeChan)
	<-b.closedChan
}
