package binder

import (
	"show-me-code/pkg/ot/text"
)

type PortalImpl struct {
	client *binderClient

	transformRcvChan <-chan text.Transform // 只读

	transformSendChan chan<- transformSubmission // 只写

	exitChan chan<- *binderClient
}

func (p *PortalImpl) SendTransform() {

}
