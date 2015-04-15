package limio

func notify(n chan<- bool, v bool) {
	if n == nil {
		return
	}

	select {
	case n <- v:
	default:
		// case <-time.After(time.Second):
	}
	if v {
		close(n)
	}
}
