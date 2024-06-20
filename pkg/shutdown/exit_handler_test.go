package shutdown

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewExitHandlerWithCtxCreatesValidLogger(t *testing.T) {
	mainCtxCanceler := func() {}
	handler := NewExitHandlerWithCtx(mainCtxCanceler)

	assert.NotNil(t, handler)
	assert.IsType(t, &exitHandler{}, handler)
	assert.NotNil(t, handler.(*exitHandler).log)
	assert.IsType(t, &zap.SugaredLogger{}, handler.(*exitHandler).log)
}

func TestAssignNonEmptySliceToCancel(t *testing.T) {
	// Create a dummy cancel function
	cancelFunc := func() {}

	// Create a slice with the dummy cancel function
	toCancel := []context.CancelFunc{cancelFunc}

	// Initialize the exitHandler
	eh := NewExitHandlerWithCtx(func() {})

	// Assign the non-empty slice to the toCancel field
	eh.ToCancel(toCancel)

	// Check if the assignment was successful
	if len(eh.(*exitHandler).toCancel) != 1 {
		t.Errorf("Expected toCancel length to be 1, got %d", len(eh.(*exitHandler).toCancel))
	}
}

func TestToStopAssignsNonEmptySlice(t *testing.T) {
	eh := &exitHandler{}
	toStop := []chan struct{}{make(chan struct{}), make(chan struct{})}
	eh.ToStop(toStop)

	if len(eh.toStop) != 2 {
		t.Errorf("expected length 2, got %d", len(eh.toStop))
	}
}

func TestToStopAssignsSliceWithNilChannels(t *testing.T) {
	eh := &exitHandler{}
	toStop := []chan struct{}{nil, nil}
	eh.ToStop(toStop)

	if len(eh.toStop) != 2 {
		t.Errorf("expected length 2, got %d", len(eh.toStop))
	}
	if eh.toStop[0] != nil || eh.toStop[1] != nil {
		t.Errorf("expected nil channels, got non-nil")
	}
}
