package handler

type Closeable interface {
	Close() error
}
