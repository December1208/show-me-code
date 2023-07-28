package text

import "bytes"

type Transform struct {
	Position  int
	Version   int
	Delete    int
	Insert    string
	TReceived int64
}

func intMin(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func intMax(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func FixTransform(current, pre *Transform) {
	currentInsert, preInsert := bytes.Runes([]byte(current.Insert)), bytes.Runes([]byte(pre.Insert))
	currentLength, preLength := len(currentInsert), len(preInsert)

	if pre.Position <= current.Position {
		if preLength > 0 && pre.Delete == 0 {
			current.Position += preLength
		} else if pre.Delete > 0 && (pre.Position+pre.Delete) <= current.Position {
			current.Position += (preLength - pre.Delete)
		} else if pre.Delete > 0 && (pre.Position+pre.Delete) > current.Position {
			overhang := intMin(current.Delete, (pre.Position+pre.Delete)-current.Position)
			current.Delete -= overhang
			current.Position = pre.Position + preLength
		}
	} else if current.Delete > 0 && (current.Position+current.Delete) > pre.Position {
		posGap := pre.Position - current.Position
		excess := intMax(0, (current.Delete - posGap))

		if excess > pre.Delete {
			current.Delete += (preLength - pre.Delete)

			newInsert := make([]rune, currentLength+preLength)
			copy(newInsert[:], currentInsert)
			copy(newInsert[currentLength:], preInsert)

			current.Insert = string(newInsert)
		} else {
			current.Delete = posGap
		}
	}
}
