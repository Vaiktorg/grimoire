package ws

import (
	"context"
	"errors"
	"nhooyr.io/websocket"
	"sync"
)

// MockWebSocketConn provides a detailed mock of the websocket.conn
type MockWebSocketConn struct {
	readCh   chan []byte     // Channel to simulate incoming messages
	writeCh  chan []byte     // Channel to simulate outgoing messages
	closeCh  chan struct{}   // Channel to simulate closing connection
	closeErr error           // Error to return on close
	readErr  error           // Error to simulate on read
	writeErr error           // Error to simulate on write
	mu       sync.Mutex      // Mutex for synchronization
	wg       *sync.WaitGroup // Mutex for synchronization
}

func NewMockConns() (w *MockWebSocketConn, r *MockWebSocketConn) {
	wg := new(sync.WaitGroup)

	w = &MockWebSocketConn{
		wg:      wg,
		readCh:  make(chan []byte, 10), // Buffered channel
		writeCh: make(chan []byte, 10),
		closeCh: make(chan struct{}),
	}

	r = &MockWebSocketConn{
		wg:      wg,
		readCh:  w.writeCh,
		writeCh: w.readCh,
		closeCh: make(chan struct{}),
	}

	return w, r
}

func (m *MockWebSocketConn) Read(ctx context.Context) (websocket.MessageType, []byte, error) {
	select {
	case <-ctx.Done():
		return 0, nil, ctx.Err()
	case <-m.closeCh:
		return 0, nil, errors.New("closed")
	case data := <-m.readCh:
		if m.readErr != nil {
			return 0, nil, m.readErr
		}

		m.wg.Done()
		return websocket.MessageText, data, nil
	}
}

func (m *MockWebSocketConn) Write(ctx context.Context, _ websocket.MessageType, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-m.closeCh:
		return errors.New("closed")
	case m.writeCh <- data:
		if m.writeErr != nil {
			return m.writeErr
		}

		m.readCh <- <-m.writeCh

		m.wg.Add(1)
		return nil
	}
}

func (m *MockWebSocketConn) Close(_ websocket.StatusCode, _ string) error {
	close(m.closeCh)

	m.wg.Wait()
	return m.closeErr
}

func (m *MockWebSocketConn) isClosed() bool {
	select {
	case <-m.closeCh:
		return true
	default:
		return false
	}
}
