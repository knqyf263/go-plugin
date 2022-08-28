package main

import (
	"testing"

	"github.com/knqyf263/go-plugin/tests"
	"github.com/stretchr/testify/assert"
)

func Test_main(t *testing.T) {
	got := tests.TestStdout(t, main)
	want := `Good morning, go-plugin
Good evening, go-plugin`
	assert.Equal(t, want, got)
}
