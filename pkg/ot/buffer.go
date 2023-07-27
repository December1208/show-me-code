package ot

import (
	"errors"
	"time"
)

var (
	TransformNegDeleteErr     = errors.New("transform contained negative delete")
	TransformPositionOOBErr   = errors.New("transform position out of bounds of document")
	TransformVersionTooOldErr = errors.New("transform version is too far behind")
	TransformSkipedErr        = errors.New("transform version greater than latest")
)

type OTBuffer struct {
	virtualLen int
	Version    int
	Applied    []Transform
	Unapplied  []Transform
}

type OTBufferConfig struct {
	MaxVirtualLen            uint64
	MaxTransformInsertLength uint64
}

func NewOTBufferConfig() OTBufferConfig {
	return OTBufferConfig{
		MaxVirtualLen:            10485760, // 10MiB
		MaxTransformInsertLength: 10240,    // 10KiB
	}
}

func (b *OTBuffer) PushTransform(ot Transform) (Transform, int, error) {
	if ot.Position < 0 {
		return Transform{}, 0, TransformPositionOOBErr
	}
	if ot.Delete < 0 {
		return Transform{}, 0, TransformNegDeleteErr
	}

	lenApplied, lenUnapplied := len(b.Applied), len(b.Unapplied)

	diff := (b.Version + 1) - ot.Version

	if diff > lenApplied+lenUnapplied {
		return Transform{}, 0, TransformVersionTooOldErr
	}
	if diff < 0 {
		return Transform{}, 0, TransformSkipedErr
	}

	for j := lenApplied - (diff - lenUnapplied); j < lenApplied; j++ {
		FixTransform(&ot, &b.Applied[j])
		diff--
	}
	for j := lenUnapplied - diff; j < lenUnapplied; j++ {
		FixTransform(&ot, &b.Unapplied[j])
	}

	if (ot.Position + ot.Delete) > b.virtualLen {
		return Transform{}, 0, TransformPositionOOBErr
	}

	b.Version++

	ot.Version = b.Version
	ot.TReceived = time.Now().Unix()

	b.Unapplied = append(b.Unapplied, ot)

	b.virtualLen += (len(ot.Insert) - ot.Delete)

	return ot, b.Version, nil
}
