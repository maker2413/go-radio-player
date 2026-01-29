package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	var s = Stack[string]{}

	assert.True(t, s.IsEmpty(), "Stack should be Empty!")

	value, err := s.Pop()
	assert.Error(t, err, "Expected error! Empty stack.")
	assert.Equal(t, "", value, "Expected nil! Empty stack.")

	s.Push("hello")
	s.Push("world")
	s.Push("!")

	assert.Len(t, s.elements, 3, "Expected 3 elements!")

	value, err = s.Pop()
	assert.NoError(t, err, "Expected no error! Stack not empty.")
	assert.Equal(t, "!", value, "Expected none empty string! Stack not empty.")
	assert.Len(t, s.elements, 2, "Expected 2 elements!")
}
