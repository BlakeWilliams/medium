package set

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	s := New[string]()
	s.Add("omg")

	require.Len(t, s.Items(), 1)
	require.True(t, s.HasItem("omg"))
	require.False(t, s.HasItem("the truth is out there"))

	s.Add("the truth is out there")

	require.Len(t, s.Items(), 2)
	require.True(t, s.HasItem("omg"))
	require.True(t, s.HasItem("the truth is out there"))

	s.Remove("omg")

	require.Len(t, s.Items(), 1)
	require.False(t, s.HasItem("omg"))
	require.True(t, s.HasItem("the truth is out there"))
}
