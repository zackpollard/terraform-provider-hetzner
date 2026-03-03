// Copyright (c) Zack
// SPDX-License-Identifier: MPL-2.0

package testutil

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestMockServer_RegisteredRoute(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.RegisterRoute(MockRoute{
		Method:     http.MethodGet,
		Path:       "/key",
		StatusCode: http.StatusOK,
		Response:   SSHKeyListResponse(SSHKeyResponse("my-key", "aa:bb:cc", "ED25519", 256, "ssh-ed25519 AAAA...")),
	})

	body, err := ms.Client().Get("/key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 key, got %d", len(result))
	}
}

func TestMockServer_UnregisteredRoute(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	_, err := ms.Client().Get("/nonexistent")
	if err == nil {
		t.Fatal("expected error for unregistered route, got nil")
	}
}

func TestMockServer_ErrorResponse(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.RegisterRoute(MockRoute{
		Method:     http.MethodGet,
		Path:       "/key/bad",
		StatusCode: http.StatusNotFound,
		Response:   ErrorResponse(404, "KEY_NOT_FOUND", "SSH key not found"),
	})

	_, err := ms.Client().Get("/key/bad")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMockServer_ClearRoutes(t *testing.T) {
	ms := NewMockServer()
	defer ms.Close()

	ms.RegisterRoute(MockRoute{
		Method:     http.MethodGet,
		Path:       "/key",
		StatusCode: http.StatusOK,
		Response:   SSHKeyListResponse(),
	})

	// Should succeed before clearing.
	_, err := ms.Client().Get("/key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ms.ClearRoutes()

	// Should fail after clearing.
	_, err = ms.Client().Get("/key")
	if err == nil {
		t.Fatal("expected error after clearing routes, got nil")
	}
}
