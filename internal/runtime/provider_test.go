package runtime

import (
	"context"
	"io"
	"testing"
)

// compile-time check: mockProvider satisfies Provider
var _ Provider = (*mockProvider)(nil)

type mockProvider struct {
	runCalled    bool
	stopCalled   bool
	removeCalled bool
	listCalled   bool
	closeCalled  bool

	runResult    string
	runErr       error
	stopErr      error
	removeErr    error
	listResult   []ContainerInfo
	listErr      error
}

func (m *mockProvider) Run(_ context.Context, _ *DesktopRunConfig, _ RunOptions, _ io.Writer) (string, error) {
	m.runCalled = true
	return m.runResult, m.runErr
}

func (m *mockProvider) Stop(_ context.Context, _ string, _ int) error {
	m.stopCalled = true
	return m.stopErr
}

func (m *mockProvider) Remove(_ context.Context, _ string, _ bool) error {
	m.removeCalled = true
	return m.removeErr
}

func (m *mockProvider) List(_ context.Context, _ bool) ([]ContainerInfo, error) {
	m.listCalled = true
	return m.listResult, m.listErr
}

func (m *mockProvider) Close() error {
	m.closeCalled = true
	return nil
}

func TestManagerDelegatesRun(t *testing.T) {
	mock := &mockProvider{runResult: "abc123"}
	mgr := NewManager(mock)

	id, err := mgr.Run(context.Background(), &DesktopRunConfig{}, RunOptions{}, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "abc123" {
		t.Errorf("expected id abc123, got %q", id)
	}
	if !mock.runCalled {
		t.Error("expected Run to be called on provider")
	}
}

func TestManagerDelegatesStop(t *testing.T) {
	mock := &mockProvider{}
	mgr := NewManager(mock)

	if err := mgr.Stop(context.Background(), "mycontainer", 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.stopCalled {
		t.Error("expected Stop to be called on provider")
	}
}

func TestManagerDelegatesRemove(t *testing.T) {
	mock := &mockProvider{}
	mgr := NewManager(mock)

	if err := mgr.Remove(context.Background(), "mycontainer", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.removeCalled {
		t.Error("expected Remove to be called on provider")
	}
}

func TestManagerDelegatesList(t *testing.T) {
	mock := &mockProvider{
		listResult: []ContainerInfo{
			{ID: "abc123", Name: "mydesk"},
		},
	}
	mgr := NewManager(mock)

	containers, err := mgr.List(context.Background(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 1 {
		t.Errorf("expected 1 container, got %d", len(containers))
	}
	if containers[0].ID != "abc123" {
		t.Errorf("expected ID abc123, got %q", containers[0].ID)
	}
	if !mock.listCalled {
		t.Error("expected List to be called on provider")
	}
}
