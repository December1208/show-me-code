package text

type OTBufferInterface interface {
	PushTransform(ot Transform) (Transform, int, error)
}
