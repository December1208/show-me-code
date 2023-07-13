package ot

import (
	"errors"
	"time"
)

type OTBuffer struct {
	virtualLen int
	Version    int
	Applied    []Transform
	Unapplied  []Transform
}

func (b *OTBuffer) PushTransform(ot Transform) (Transform, int, error) {
	if ot.Position < 0 {
		return Transform{}, 0, errors.New("123")
	}
	if ot.Delete < 0 {
		return Transform{}, 0, errors.New("123")
	}
	// if uint64(len(ot.Insert)) > b.config.MaxTransformLength {
	// 	return Transform{}, 0, errors.New("123")
	// }

	lenApplied, lenUnapplied := len(b.Applied), len(b.Unapplied)

	diff := (b.Version + 1) - ot.Version

	if diff > lenApplied {
		return Transform{}, 0, errors.New("123")
	}
	if diff < 0 {
		return Transform{}, 0, errors.New("123")
	}

	for j := lenApplied - (diff - lenUnapplied); j < lenApplied; j++ {
		FixTransform(&ot, &b.Applied[j])
		diff--
	}
	for j := lenUnapplied - diff; j < lenUnapplied; j++ {
		FixTransform(&ot, &b.Unapplied[j])
	}

	// After adjustment check for document size bounds.
	// if uint64(len(ot.Insert)-ot.Delete+b.virtualLen) > b.config.MaxDocumentSize {
	// 	return Transform{}, 0, errors.New("123")
	// }
	if (ot.Position + ot.Delete) > b.virtualLen {
		return Transform{}, 0, errors.New("123")
	}

	b.Version++

	ot.Version = b.Version
	ot.TReceived = time.Now().Unix()

	b.Unapplied = append(b.Unapplied, ot)

	b.virtualLen += (len(ot.Insert) - ot.Delete)

	return ot, b.Version, nil
}
