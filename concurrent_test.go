package codex

import (
	"sync"
	"testing"
)

// TestThreadIDConcurrentRead tests that Thread.ID() is safe to call concurrently.
func TestThreadIDConcurrentRead(t *testing.T) {
	client, err := New()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	thread := client.StartThread()

	// Set an ID first
	thread.setID("test-thread-id")

	// Concurrently read the ID from multiple goroutines
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			id := thread.ID()
			if id != "test-thread-id" {
				t.Errorf("expected ID %q, got %q", "test-thread-id", id)
			}
		}()
	}

	wg.Wait()
}

// TestThreadSetIDConcurrent tests that Thread.setID() handles concurrent calls safely.
func TestThreadSetIDConcurrent(t *testing.T) {
	client, err := New()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	thread := client.StartThread()

	// Concurrently set IDs from multiple goroutines
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			// All goroutines try to set the same ID
			// This tests that the mutex properly serializes access
			thread.setID("concurrent-id")
		}()
	}

	wg.Wait()

	// Verify the ID was set
	if thread.ID() == "" {
		t.Error("expected ID to be set")
	}
}

// TestStreamedTurnWaitConcurrent tests that StreamedTurn.Wait() is safe to call concurrently.
func TestStreamedTurnWaitConcurrent(t *testing.T) {
	// Create a StreamedTurn with a simple wait function
	called := 0
	var mu sync.Mutex

	streamed := &StreamedTurn{
		Events: make(<-chan ThreadEvent),
		waitFn: func() error {
			mu.Lock()
			called++
			mu.Unlock()
			return nil
		},
	}

	// Concurrently call Wait from multiple goroutines
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			err := streamed.Wait()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}

	wg.Wait()

	// Verify that waitFn was called exactly once
	mu.Lock()
	defer mu.Unlock()
	if called != 1 {
		t.Errorf("expected waitFn to be called once, got %d calls", called)
	}
}

// TestThreadConcurrentOperations tests multiple concurrent operations on a Thread.
func TestThreadConcurrentOperations(t *testing.T) {
	client, err := New()
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	thread := client.StartThread()
	thread.setID("test-id")

	// Mix of concurrent reads and writes
	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Readers
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = thread.ID()
		}()
	}

	// Writers (setting the same ID repeatedly)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			thread.setID("test-id")
		}()
	}

	wg.Wait()

	// Verify final state
	if thread.ID() != "test-id" {
		t.Errorf("expected ID %q, got %q", "test-id", thread.ID())
	}
}

// TestExecStreamConcurrentClose tests that ExecStream.Close() is safe to call concurrently.
func TestExecStreamConcurrentClose(t *testing.T) {
	// Create a mock stdout
	r, w := newPipe()
	defer w.Close()

	stream := &ExecStream{
		stdout: r,
		waitFn: func() error { return nil },
	}

	// Concurrently close from multiple goroutines
	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = stream.Close()
		}()
	}

	wg.Wait()
}

// newPipe creates a simple pipe for testing.
func newPipe() (*pipeReader, *pipeWriter) {
	pr := &pipeReader{closed: make(chan struct{})}
	pw := &pipeWriter{pr: pr}
	return pr, pw
}

type pipeReader struct {
	closed chan struct{}
	once   sync.Once
}

func (r *pipeReader) Read(p []byte) (n int, err error) {
	<-r.closed
	return 0, nil
}

func (r *pipeReader) Close() error {
	r.once.Do(func() {
		close(r.closed)
	})
	return nil
}

type pipeWriter struct {
	pr *pipeReader
}

func (w *pipeWriter) Close() error {
	return w.pr.Close()
}
