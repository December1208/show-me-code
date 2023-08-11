package text

type Type interface {
	PushTransform(ot Transform) (Transform, int, error)
}
