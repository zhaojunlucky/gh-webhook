package model

type Closeable interface {
	Close() error
}
