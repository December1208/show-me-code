package store

import "show-me-code/pkg/util"

type Document struct {
	ID      string
	Content string
}

func NewDocument(content string) Document {
	return Document{
		ID:      util.GenerateUUID(),
		Content: content,
	}
}
