package stack

import (
	"errors"
	"sync"
)

type Stack[T any] struct {
	lock     sync.Mutex
	elements []T
}

func (s *Stack[T]) Push(value T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.elements = append(s.elements, value)
}

func (s *Stack[T]) Pop() (T, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.IsEmpty() {
		var zero T
		return zero, errors.New("Stack is currently empty! Can not pop!")
	}

	index := len(s.elements) - 1
	value := s.elements[index]
	s.elements = s.elements[:index]

	return value, nil
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.elements) == 0
}
