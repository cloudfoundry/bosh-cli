package disk

func withinDelta(left, right, delta uint64) bool {
	switch {
	case left-delta >= right:
		return true
	case right-delta <= left:
		return true
	}
	return false
}
