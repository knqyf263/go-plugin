package tests

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	t.Cleanup(func() {
		os.Stdout = old
	})
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = w
	fn()
	w.Close()

	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(r)
	require.NoError(t, err)

	s := buffer.String()
	return s[:len(s)-1]
}
