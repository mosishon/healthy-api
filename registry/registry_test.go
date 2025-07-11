package registry_test

import (
	"healthy-api/model"
	"healthy-api/notifier"
	"healthy-api/registry"
	"testing"
)

type RegistryTestNotifier struct {
	ID string
}

func (m *RegistryTestNotifier) Notify(n model.Notification) error {
	return nil // No logic needed for this test
}
func (m *RegistryTestNotifier) GetName() string {
	return "Test name"
}

func TestRegistry(t *testing.T) {
	registry := registry.NewRegistry[notifier.Notifier]()
	notifier1 := &RegistryTestNotifier{ID: "notifier1"}
	notifier2 := &RegistryTestNotifier{ID: "notifier2"}
	notifierID1 := "email-alerts"
	notifierID2 := "sms-alerts"

	registry.Register(notifierID1, notifier1)
	registry.Register(notifierID2, notifier2)

	// Case 1: Retrieve a notifier that exists
	retrieved1, ok := registry.Get(notifierID1)
	if ok == false {
		t.Fatalf("Expected to retrieve notifier '%s', but got nil", notifierID1)
	}
	if retrieved1 != notifier1 {
		t.Errorf("Retrieved notifier is not the same instance as the registered one")
	}

	// Case 2: Retrieve the other notifier
	retrieved2, ok := registry.Get(notifierID2)
	if ok == false {
		t.Fatalf("Expected to retrieve notifier '%s', but got nil", notifierID2)
	}
	if retrieved2 != notifier2 {
		t.Errorf("Retrieved notifier is not the same instance as the registered one")
	}

	// Case 3: Retrieve a notifier that does NOT exist
	_, ok = registry.Get("non-existent-notifier")
	if ok == true {
		t.Errorf("Expected Get to return nil for a non-existent notifier, but it returned an object")
	}
}
