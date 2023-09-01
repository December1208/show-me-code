package binder

import "show-me-code/pkg/ot/text"

type Error struct {
	ID  string
	Err error
}

type binderClient struct {
	metadata interface{}

	transformSendChan chan<- text.Transform
}

type transformSubmission struct {
	client    *binderClient
	transform text.Transform
}
