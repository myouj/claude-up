package worker

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSSEManager_Subscribe(t *testing.T) {
	mgr := NewSSEManager()

	ch := mgr.Subscribe(1)
	if ch == nil {
		t.Error("Subscribe should return a channel")
	}

	// Verify the channel is stored
	mgr.mu.RLock()
	stored, ok := mgr.clients[1]
	mgr.mu.RUnlock()

	if !ok {
		t.Error("Client channel not stored in manager")
	}
	if stored != ch {
		t.Error("Stored channel different from returned channel")
	}
}

func TestSSEManager_Unsubscribe(t *testing.T) {
	mgr := NewSSEManager()

	ch := mgr.Subscribe(1)
	mgr.Unsubscribe(1)

	// Verify the channel is removed
	mgr.mu.RLock()
	_, ok := mgr.clients[1]
	mgr.mu.RUnlock()

	if ok {
		t.Error("Client channel should be removed after Unsubscribe")
	}

	// Verify channel is closed
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("Channel should be closed after Unsubscribe")
		}
	default:
		t.Error("Channel should be closed and empty after Unsubscribe")
	}
}

func TestSSEManager_Publish(t *testing.T) {
	mgr := NewSSEManager()

	// Subscribe
	ch := mgr.Subscribe(1)

	// Publish an event
	event := SSEEvent{
		Event: "test",
		Data:  map[string]string{"message": "hello"},
	}
	mgr.Publish(1, event)

	// Verify event is received
	select {
	case received := <-ch:
		if received.Event != event.Event {
			t.Errorf("Expected event '%s', got '%s'", event.Event, received.Event)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Event was not received on channel")
	}
}

func TestSSEManager_Publish_NoSubscriber(t *testing.T) {
	mgr := NewSSEManager()

	// Should not panic when publishing to non-existent client
	event := SSEEvent{
		Event: "test",
		Data:  map[string]string{"message": "hello"},
	}
	mgr.Publish(999, event) // No subscriber for ID 999
}

func TestSSEManager_SendSSEProgress(t *testing.T) {
	mgr := NewSSEManager()
	ch := mgr.Subscribe(1)

	mgr.SendSSEProgress(1, 5, 20, "running")

	select {
	case received := <-ch:
		if received.Event != "progress" {
			t.Errorf("Expected event 'progress', got '%s'", received.Event)
		}

		data, ok := received.Data.(SSEProgress)
		if !ok {
			t.Fatal("Data should be SSEProgress type")
		}
		if data.Current != 5 {
			t.Errorf("Expected current 5, got %d", data.Current)
		}
		if data.Total != 20 {
			t.Errorf("Expected total 20, got %d", data.Total)
		}
		if data.Progress != 25 {
			t.Errorf("Expected progress 25, got %d", data.Progress)
		}
		if data.Status != "running" {
			t.Errorf("Expected status 'running', got '%s'", data.Status)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Event was not received on channel")
	}
}

func TestSSEManager_SendSSEProgress_ZeroTotal(t *testing.T) {
	mgr := NewSSEManager()
	ch := mgr.Subscribe(1)

	mgr.SendSSEProgress(1, 0, 0, "running")

	select {
	case received := <-ch:
		data, ok := received.Data.(SSEProgress)
		if !ok {
			t.Fatal("Data should be SSEProgress type")
		}
		if data.Progress != 0 {
			t.Errorf("Expected progress 0 when total is 0, got %d", data.Progress)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Event was not received on channel")
	}
}

func TestSSEManager_SendSSEComplete(t *testing.T) {
	mgr := NewSSEManager()
	ch := mgr.Subscribe(1)

	result := map[string]interface{}{"message": "success", "count": 42}
	mgr.SendSSEComplete(1, result)

	select {
	case received := <-ch:
		if received.Event != "complete" {
			t.Errorf("Expected event 'complete', got '%s'", received.Event)
		}

		data, ok := received.Data.(SSEComplete)
		if !ok {
			t.Fatal("Data should be SSEComplete type")
		}
		if data.Status != "done" {
			t.Errorf("Expected status 'done', got '%s'", data.Status)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Event was not received on channel")
	}
}

func TestSSEManager_SendSSEError(t *testing.T) {
	mgr := NewSSEManager()
	ch := mgr.Subscribe(1)

	errMsg := "something went wrong"
	mgr.SendSSEError(1, errMsg)

	select {
	case received := <-ch:
		if received.Event != "error" {
			t.Errorf("Expected event 'error', got '%s'", received.Event)
		}

		data, ok := received.Data.(SSEError)
		if !ok {
			t.Fatal("Data should be SSEError type")
		}
		if data.Status != "failed" {
			t.Errorf("Expected status 'failed', got '%s'", data.Status)
		}
		if data.Error != errMsg {
			t.Errorf("Expected error '%s', got '%s'", errMsg, data.Error)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Event was not received on channel")
	}
}

func TestSSEEvent_JSON(t *testing.T) {
	event := SSEEvent{
		Event: "progress",
		Data: SSEProgress{
			Current:  5,
			Total:    20,
			Progress: 25,
			Status:   "running",
		},
	}

	jsonBytes, err := json.Marshal(event.Data)
	if err != nil {
		t.Fatalf("Failed to marshal SSEProgress: %v", err)
	}

	var decoded SSEProgress
	err = json.Unmarshal(jsonBytes, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal SSEProgress: %v", err)
	}

	if decoded.Current != 5 {
		t.Errorf("Expected current 5, got %d", decoded.Current)
	}
	if decoded.Total != 20 {
		t.Errorf("Expected total 20, got %d", decoded.Total)
	}
	if decoded.Progress != 25 {
		t.Errorf("Expected progress 25, got %d", decoded.Progress)
	}
	if decoded.Status != "running" {
		t.Errorf("Expected status 'running', got '%s'", decoded.Status)
	}
}
