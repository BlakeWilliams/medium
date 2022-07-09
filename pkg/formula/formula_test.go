package formula

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Name       string
	ID         int
	Balance    float32
	Int8       int8
	Int16      int16
	Int32      int32
	Int64      int64
	Int        int
	Uint8      uint8
	Uint16     uint16
	Uint32     uint32
	Uint64     uint64
	Uint       uint
	Complex64  complex64
	Complex128 complex128
	Foods      []string
	Truthy     bool
	private    string
}

func TestDecode(t *testing.T) {
	target := testStruct{}
	data := map[string][]string{
		"Name":       {"Fox Mulder"},
		"ID":         {"1"},
		"Balance":    {"53.3"},
		"Int8":       {"8"},
		"Int16":      {"16"},
		"Int32":      {"32"},
		"Int64":      {"64"},
		"Int":        {"100"},
		"Uint8":      {"8"},
		"Uint16":     {"16"},
		"Uint32":     {"32"},
		"Uint64":     {"64"},
		"Uint":       {"1000"},
		"Complex64":  {"64.5"},
		"Complex128": {"128.5"},
		"Foods":      {"pizza", "hotdogs"},
		"Truthy":     {"true"},
		"private":    {"omg"},
	}

	decoder := Decoder{}
	err := decoder.Decode(&target, data)
	require.NoError(t, err)

	require.Equal(t, "Fox Mulder", target.Name)
	require.Equal(t, 1, target.ID)
	require.Equal(t, float32(53.3), target.Balance)
	require.Equal(t, int8(8), target.Int8)
	require.Equal(t, int16(16), target.Int16)
	require.Equal(t, int32(32), target.Int32)
	require.Equal(t, int64(64), target.Int64)
	require.Equal(t, int(100), target.Int)
	require.Equal(t, uint8(8), target.Uint8)
	require.Equal(t, uint16(16), target.Uint16)
	require.Equal(t, uint32(32), target.Uint32)
	require.Equal(t, uint64(64), target.Uint64)
	require.Equal(t, uint(1000), target.Uint)
	require.Equal(t, complex64(64.5), target.Complex64)
	require.Equal(t, complex128(128.5), target.Complex128)
	require.Equal(t, []string{"pizza", "hotdogs"}, target.Foods)
	require.True(t, target.Truthy)
	require.Equal(t, "", target.private)
}

func TestNil(t *testing.T) {
	target := testStruct{}
	data := map[string][]string{}

	decoder := Decoder{}
	err := decoder.Decode(&target, data)
	require.NoError(t, err)

	require.Zero(t, target.Name)
	require.Zero(t, target.ID)
	require.Zero(t, target.Balance)
	require.Zero(t, target.Foods)
}

func TestFailedDecoding(t *testing.T) {
	target := testStruct{}
	data := map[string][]string{"Uint32": {"lol"}}

	decoder := Decoder{}
	err := decoder.Decode(&target, data)

	require.Error(t, err)
}

func TestDecodePointer(t *testing.T) {
	target := struct {
		Int     *int
		Numbers *[]*int
	}{}

	data := map[string][]string{
		"Int":     {"5"},
		"Numbers": {"5", "2"},
	}

	decoder := Decoder{}
	err := decoder.Decode(&target, data)
	require.NoError(t, err)

	myInt := new(int)
	*myInt = 5

	otherInt := new(int)
	*otherInt = 2
	require.Equal(t, myInt, target.Int)

	slice := []*int{myInt, otherInt}
	require.Equal(t, &slice, target.Numbers)
}

func TestIgnoresDoublePointers(t *testing.T) {
	target := struct {
		Int     **int
		Numbers **[]*int
	}{}

	data := map[string][]string{
		"Int":     {"5"},
		"Numbers": {"5", "2"},
	}

	decoder := Decoder{}
	err := decoder.Decode(&target, data)
	require.NoError(t, err)
}

func TestCustomNames(t *testing.T) {
	target := struct {
		Name string `param:"my_name"`
	}{}

	data := map[string][]string{
		"my_name": {"Fox Mulder"},
	}

	decoder := Decoder{}
	err := decoder.Decode(&target, data)
	require.NoError(t, err)

	require.Equal(t, target.Name, "Fox Mulder")
}

func TestInvalidTypes(t *testing.T) {
	target := struct {
		Name string `param:"my_name"`
	}{}
	refTarget := &target

	data := map[string][]string{
		"my_name": {"Fox Mulder"},
	}

	decoder := Decoder{}
	require.Error(t, decoder.Decode(target, data))
	require.Error(t, decoder.Decode(nil, data))
	require.Error(t, decoder.Decode(&refTarget, data))
	require.Error(t, decoder.Decode(1, data))
	require.Error(t, decoder.Decode("omg", data))
}
