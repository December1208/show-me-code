package binder

import "show-me-code/pkg/ot/text"

type ClientMetadata struct {
	Client   interface{}
	Metadata interface{}
}

type BinderClient struct {
	transformChan chan<- text.Transform
	metaDataChan  chan<- ClientMetadata
}

type Portal struct {
	client *BinderClient

	// exitChan chan<-
}
