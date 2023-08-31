package binder

import "show-me-code/pkg/ot/text"

type Error struct {
	ID  string
	Err error
}

type ClientMetadata struct {
	Client   interface{}
	Metadata interface{}
}

type binderClient struct {
	metadata interface{}

	transformSendChan chan<- text.Transform
}

type transformSubmission struct {
	client    *binderClient
	transform text.Transform
}
