package page

const Size = 8192

type Type uint8

const (
	Pointer Type = iota + 1
	Leaf
	Free
)

func GetType(page [Size]byte) Type {
	return Type(page[0])
}
