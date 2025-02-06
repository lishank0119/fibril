package fibril

type filterFunc func(*Client) bool

type box struct {
	t      int
	msg    []byte
	filter filterFunc
}
