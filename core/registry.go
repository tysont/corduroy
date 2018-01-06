package corduroy

type Registry interface {
	Put(id int, address string)
	Get(id int) string
	GetIDs(start int, length int) []int
	GetRandomID() int
	GetAll() map[int]string
	Delete(id int)
	Contains(id int) bool
	Size() int
}
