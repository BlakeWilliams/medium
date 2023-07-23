package set

// Set represents the Set datatype, which is a collection of unique items that
// can be iterated over.
//
// The order of the set is not guaranteed and is not safe for concurrent access.
type Set[T comparable] struct {
	backing map[T]struct{}
}

// Creates a new set of type T
func New[T comparable]() *Set[T] {
	return &Set[T]{backing: make(map[T]struct{})}
}

// Adds item to set. The item is only added once based on it's `comparable`
// implementation.
func (s *Set[T]) Add(item T) {
	s.backing[item] = struct{}{}
}

// Checks if the item exists in the set.
func (s *Set[T]) HasItem(item T) bool {
	_, ok := s.backing[item]
	return ok
}

// Removes the given item from the set.
func (s *Set[T]) Remove(item T) {
	delete(s.backing, item)
}

// Returns the number of elements in the set.
func (s *Set[T]) Len() int {
	return len(s.backing)
}

// Returns a slice of items in the set. Modifying the slice will not modify the
// set.
func (s *Set[T]) Items() []T {
	items := make([]T, 0, s.Len())
	for item := range s.backing {
		items = append(items, item)
	}
	return items
}
