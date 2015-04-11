package limio

func (r *Reader) Close() error {
	r.Unlimit()
	return nil
}

func (r *Reader) Read(p []byte) (n int, err error) {
	return
}
