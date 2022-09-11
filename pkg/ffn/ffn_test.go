package ffn

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	collection := []int{1, 2, 3}

	res := Map(collection, func(i int) int {
		return i * 2
	})

	require.Equal(t, []int{2, 4, 6}, res)
}

func ExampleMap() {
	collection := []int{1, 2, 3}

	res := Map(collection, func(i int) int {
		return i * 2
	})

	fmt.Println(res)

	// Output: [2 4 6]
}

func TestSelect(t *testing.T) {
	collection := []int{1, 2, 3}

	res := Select(collection, func(i int) bool {
		return i%2 == 0
	})

	require.Equal(t, []int{2}, res)
}

func ExampleSelect() {
	collection := []int{1, 2, 3, 4}

	res := Select(collection, func(i int) bool {
		return i%2 == 0
	})

	fmt.Println(res)

	// Output: [2 4]
}

func TestReject(t *testing.T) {
	collection := []int{1, 2, 3}

	res := Reject(collection, func(i int) bool {
		return i%2 == 0
	})

	require.Equal(t, []int{1, 3}, res)
}

func ExampleReject() {
	collection := []int{1, 2, 3, 4}

	res := Reject(collection, func(i int) bool {
		return i%2 == 0
	})

	fmt.Println(res)

	// Output: [1 3]
}

func TestReduce(t *testing.T) {
	collection := []int{1, 2, 3}

	res := Reduce(collection, func(i, acc int) int {
		return acc + i
	}, 0)

	require.Equal(t, 6, res)
}

func ExampleReduce() {
	collection := []int{1, 2, 3}

	res := Reduce(collection, func(i, acc int) int {
		return acc + i
	}, 0)

	fmt.Println(res)

	// Output: 6
}

func TestKeyBy(t *testing.T) {
	collection := []string{"first", "second", "last"}

	res := KeyBy(collection, func(s string) int {
		return len(s)
	})

	require.Equal(t, map[int]string{5: "first", 6: "second", 4: "last"}, res)
}

func ExampleKeyBy() {
	collection := []string{"first", "second", "last"}

	res := KeyBy(collection, func(s string) int {
		return len(s)
	})

	fmt.Println(res)

	// Unordered output: map[4:last 5:first 6:second]
}

func TestAll(t *testing.T) {
	collection := []int{1, 2, 3}

	res := All(collection, func(i int) bool {
		return i > 0
	})

	require.True(t, res)
}

func TestAll_False(t *testing.T) {
	collection := []int{1, 2, 3, -1}

	res := All(collection, func(i int) bool {
		return i > 0
	})

	require.False(t, res)
}

func ExampleAll() {
	collection := []int{1, 2, 3}

	res := All(collection, func(i int) bool {
		return i > 0
	})

	fmt.Println(res)

	// Output: true
}
