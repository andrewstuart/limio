package limio

func notify(n chan<- bool, v bool) {
	if n == nil {
		return
	}

	n <- v
	if v {
		close(n)
	}
}
