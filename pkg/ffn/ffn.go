package ffn

// Map calls mapper on each element on collection, returning a new collection
// with the results of each call to mapper.
func Map[T any, R any](collection []T, mapper func(T) R) []R {
	res := make([]R, len(collection))

	for i, v := range collection {
		res[i] = mapper(v)
	}

	return res
}

// Select calls selecter on each element of collection, returning a new slice for all
// elements where f returns true.
func Select[T any](collection []T, selecter func(T) bool) []T {
	res := make([]T, 0)

	for _, v := range collection {
		if selecter(v) {
			res = append(res, v)
		}
	}

	return res
}

// Reject calls rejecter on each element of collection, returning a new slice for all
// elements where f returns false.
func Reject[T any](collection []T, rejecter func(T) bool) []T {
	res := make([]T, 0)

	for _, v := range collection {
		if !rejecter(v) {
			res = append(res, v)
		}
	}

	return res
}

// Reduce accepts a slice of values, a function that accepts an element of the
// slice and an accumulater, and an initial value for the accumulator. The
// result of each call to reducer is passed as the value to the next iteration.
func Reduce[T, R any](collection []T, reducer func(T, R) R, initial R) R {
	res := initial

	for _, v := range collection {
		res = reducer(v, res)
	}

	return res
}

// KeyBy calls keygen on each element of collection, setting the result of f as the
// key in the returned map, and the element as the value.
func KeyBy[T any, K comparable](collection []T, keygen func(T) K) map[K]T {
	res := make(map[K]T, len(collection))

	for _, v := range collection {
		res[keygen(v)] = v
	}

	return res
}

// All returns true if all elements in collection return true when passed to f,
// otherwise it returns false.
func All[T any](collection []T, f func(T) bool) bool {
	for _, v := range collection {
		if !f(v) {
			return false
		}
	}

	return true
}
