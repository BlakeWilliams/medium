package formula

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Name       string     `param:"name"`
	ID         int        `param:"id"`
	Balance    float32    `param:"balance"`
	Int8       int8       `param:"int8"`
	Int16      int16      `param:"int16"`
	Int32      int32      `param:"int32"`
	Int64      int64      `param:"int64"`
	Int        int        `param:"int"`
	Uint8      uint8      `param:"uint8"`
	Uint16     uint16     `param:"uint16"`
	Uint32     uint32     `param:"uint32"`
	Uint64     uint64     `param:"uint64"`
	Uint       uint       `param:"uint"`
	Complex64  complex64  `param:"complex64"`
	Complex128 complex128 `param:"complex128"`
	Foods      []string   `param:"foods"`
	Truthy     bool       `param:"truthy"`
	private    string     `param:"private"`
}

func TestDecode(t *testing.T) {
	target := testStruct{}
	data := map[string][]string{
		"name":       {"Fox Mulder"},
		"id":         {"1"},
		"balance":    {"53.3"},
		"int8":       {"8"},
		"int16":      {"16"},
		"int32":      {"32"},
		"int64":      {"64"},
		"int":        {"100"},
		"uint8":      {"8"},
		"uint16":     {"16"},
		"uint32":     {"32"},
		"uint64":     {"64"},
		"uint":       {"1000"},
		"complex64":  {"64.5"},
		"complex128": {"128.5"},
		"foods":      {"pizza", "hotdogs"},
		"truthy":     {"true"},
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
	data := map[string][]string{"uint32": {"lol"}}

	decoder := Decoder{}
	err := decoder.Decode(&target, data)

	require.Error(t, err)
}

func TestDecodePointer(t *testing.T) {
	target := struct {
		Int     *int    `param:"int"`
		Numbers *[]*int `param:"numbers"`
	}{}

	data := map[string][]string{
		"int":     {"5"},
		"numbers": {"5", "2"},
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
		Int     **int`param:"int"`
		Numbers **[]*int `param:"numbers"`
	}{}

	data := map[string][]string{
		"int":     {"5"},
		"numbers": {"5", "2"},
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
