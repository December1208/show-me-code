package store

type Store interface {
	Create(document Document) error
	Update(document Document) error
	Read(id string) (Document, error)
}
