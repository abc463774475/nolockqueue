package nolockqueue

import (
	_ "github.com/abc463774475/my_tool/n_log"
	nlog "github.com/abc463774475/my_tool/n_log"
	"log"
	"testing"
)

func TestNewQueue(t *testing.T) {
	log.Default()
	queue := NewQueue(1)
	if queue.Len() != 1 {
		t.Errorf("Expected len 1, got %d", queue.Len())
		return
	}

	if queue.Value() != 1 {
		t.Errorf("Expected value 1, got %d", queue.Value())
		return
	}

	queue.Push(2)

	if queue.Len() != 2 {
		t.Errorf("Expected len 2, got %d", queue.Len())
		return
	}

	queue.Push(3)

	if queue.Len() != 3 {
		t.Errorf("Expected len 3, got %d", queue.Len())
		return
	}

	queue.Do(func(value interface{}) {
		nlog.Info("queue value %d", value)
	})

	nlog.Info("queue len %d", queue.Len())

	data := queue.Pop()
	nlog.Info("queue len %d val %v", queue.Len(), data)

	queue.Do(func(value interface{}) {
		nlog.Info("queue value %d", value)
	})
}
