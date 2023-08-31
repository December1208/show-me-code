package curator

import (
	"show-me-code/pkg/ot/binder"
	"show-me-code/pkg/ot/store"
	"show-me-code/pkg/util"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Config struct {
	BinderConfig binder.Config
}

type CuratorImpl struct {
	store       store.Type
	openBinder  map[string]binder.Type
	binderMutex sync.Mutex
	config      Config

	// Control channels
	errorChan  chan binder.Error
	closeChan  chan struct{}
	closedChan chan struct{}
}

func NewCuratorImpl(
	store store.Type,
) (*CuratorImpl, error) {
	curator := CuratorImpl{
		store:      store,
		openBinder: make(map[string]binder.Type),
		errorChan:  make(chan binder.Error),
		closeChan:  make(chan struct{}),
		closedChan: make(chan struct{}),
	}

	go curator.loop()
	return &curator, nil
}

func (impl *CuratorImpl) loop() {
	for {
		select {
		case err := <-impl.errorChan:
			if err.Err != nil {
				util.Logger.Error("binder err", zap.String("binderId", err.ID), zap.String("error", err.Err.Error()))
			} else {
				util.Logger.Info("binder shutdown", zap.String("binderId", err.ID))
			}
			impl.binderMutex.Lock()
			if b, ok := impl.openBinder[err.ID]; ok {
				b.Close()
				delete(impl.openBinder, err.ID)
				util.Logger.Info("close binder suc", zap.String("binderId", err.ID))

			}
			impl.binderMutex.Unlock()
		case <-impl.closeChan:
			impl.binderMutex.Lock()
			for _, b := range impl.openBinder {
				b.Close()
			}
			util.Logger.Info("close all binder suc")
			impl.binderMutex.Unlock()
			close(impl.closedChan)
			return
		}
	}
}

func (impl *CuratorImpl) Close() {
	util.Logger.Info("close curator")
	impl.closeChan <- struct{}{}
	<-impl.closedChan
}

func (impl *CuratorImpl) newBinder(id string) (binder.Type, error) {

	return binder.NewBinder(id, impl.store, impl.config.BinderConfig, impl.errorChan)
}

func (impl *CuratorImpl) CreateDocument(doc store.Document, timeout time.Duration) (binder.Portal, error) {
	if err := impl.store.Create(doc); err != nil {
		return nil, err
	}

	return nil, nil
}

func (impl *CuratorImpl) EditDocument(documentId string, timeout time.Duration) (binder.Portal, error) {
	impl.binderMutex.Lock()
	if binder, ok := impl.openBinder[documentId]; ok {
		impl.binderMutex.Unlock()
		return binder, nil
	}
	binder, err := impl.newBinder(documentId)
	if err != nil {
		impl.binderMutex.Unlock()
		return nil, err
	}

	impl.openBinder[documentId] = binder
	impl.binderMutex.Unlock()
	return nil, nil
}
