package observer

type Block[T any] struct {
	Channel string
	Block   T
}
