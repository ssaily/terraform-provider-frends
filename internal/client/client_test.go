package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer starts a mock HTTP server and returns it together with a Client
// pre-configured to talk to it.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv, NewClient(srv.URL, "test-token")
}

// writeJSON encodes v as JSON into w with status 200.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// --- Authorization ---

func TestClient_authorizationHeaderSent(t *testing.T) {
	t.Parallel()
	var got string
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
		writeJSON(w, EnvironmentListResponse{Data: []EnvironmentGet{}})
	})

	if _, err := c.ListEnvironments(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Bearer test-token" {
		t.Fatalf("expected Authorization: Bearer test-token, got %q", got)
	}
}

// --- Environment Variables ---

func TestGetEnvironmentVariable_found(t *testing.T) {
	t.Parallel()
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/environment-variables/7" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, EnvironmentVariableSchemaGet{ID: 7, Name: "DB.Host", Type: "String"})
	})

	ev, err := c.GetEnvironmentVariable(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev == nil {
		t.Fatal("expected non-nil result")
	}
	if ev.Name != "DB.Host" || ev.ID != 7 || ev.Type != "String" {
		t.Fatalf("unexpected result: %+v", ev)
	}
}

func TestGetEnvironmentVariable_notFound(t *testing.T) {
	t.Parallel()
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	ev, err := c.GetEnvironmentVariable(context.Background(), 999)
	if err != nil {
		t.Fatalf("expected nil error on 404, got: %v", err)
	}
	if ev != nil {
		t.Fatalf("expected nil result on 404, got: %+v", ev)
	}
}

func TestCreateEnvironmentVariable_success(t *testing.T) {
	t.Parallel()
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/environment-variables" {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		var req EnvironmentVariableSchemaCreate
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, EnvironmentVariableSchemaGet{ID: 42, Name: req.Name, Type: "String"})
	})

	result, err := c.CreateEnvironmentVariable(context.Background(), EnvironmentVariableSchemaCreate{Name: "Svc.Url"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 42 || result.Name != "Svc.Url" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestDeleteEnvironmentVariable_success(t *testing.T) {
	t.Parallel()
	var called bool
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == "/api/v1/environment-variables/5" {
			called = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	})

	if err := c.DeleteEnvironmentVariable(context.Background(), 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("DELETE endpoint was not called")
	}
}

// --- API Error handling ---

func TestDoRequest_apiError(t *testing.T) {
	t.Parallel()
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"detail":"forbidden"}`, http.StatusForbidden)
	})

	_, err := c.GetEnvironmentVariable(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Fatalf("expected 403 in error, got: %v", err)
	}
}

// --- Processes ---

func TestGetProcess_found(t *testing.T) {
	t.Parallel()
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/processes/10" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, ProcessGet{
			ID:               10,
			Name:             "OrderProcessor",
			UniqueIdentifier: "aaaaaaaa-0000-0000-0000-000000000001",
			Version:          3,
		})
	})

	p, err := c.GetProcess(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil || p.Name != "OrderProcessor" || p.Version != 3 {
		t.Fatalf("unexpected result: %+v", p)
	}
}

func TestGetProcess_notFound(t *testing.T) {
	t.Parallel()
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	p, err := c.GetProcess(context.Background(), 99)
	if err != nil {
		t.Fatalf("expected nil error on 404, got: %v", err)
	}
	if p != nil {
		t.Fatalf("expected nil result on 404")
	}
}

func TestDeleteProcess_success(t *testing.T) {
	t.Parallel()
	var called bool
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/processes/abc-guid") {
			called = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	})

	if err := c.DeleteProcess(context.Background(), "abc-guid"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("DELETE endpoint was not called")
	}
}

// --- Private Applications ---

func TestGetPrivateApplication_found(t *testing.T) {
	t.Parallel()
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/private-application/3" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, PrivateApplicationGet{
			ID:                       3,
			Name:                     "OrderService",
			DefaultTokenLifetimeDays: 30,
		})
	})

	app, err := c.GetPrivateApplication(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app == nil || app.Name != "OrderService" {
		t.Fatalf("unexpected result: %+v", app)
	}
}

// --- ExportProcess ---

func TestExportProcess_returnsBytes(t *testing.T) {
	t.Parallel()
	wantBody := []byte(`{"name":"OrderProcessor","version":3}`)
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/processes/10/export" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(wantBody)
	})

	got, err := c.ExportProcess(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(wantBody) {
		t.Fatalf("expected body %q, got %q", wantBody, got)
	}
}

func TestExportProcess_notFound(t *testing.T) {
	t.Parallel()
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	got, err := c.ExportProcess(context.Background(), 999)
	if err != nil {
		t.Fatalf("expected nil error on 404, got: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil bytes on 404, got %q", got)
	}
}

// --- API Policies ---

func TestCreateApiPolicy_sendsCorrectBody(t *testing.T) {
	t.Parallel()
	var received ApiPolicySave
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/api-policies" {
			http.Error(w, "unexpected", http.StatusBadRequest)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, ApiPolicyGet{ID: 1, Name: received.Name})
	})

	_, err := c.CreateApiPolicy(context.Background(), ApiPolicySave{
		Name:              "MyPolicy",
		AllowPublicAccess: false,
		TargetEndpoints:   []ApiPolicyTargetEndpointSave{{URL: "https://example.com", Method: "GET"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Name != "MyPolicy" {
		t.Fatalf("expected Name=MyPolicy in request body, got %q", received.Name)
	}
	if len(received.TargetEndpoints) != 1 || received.TargetEndpoints[0].URL != "https://example.com" {
		t.Fatalf("unexpected target endpoints: %+v", received.TargetEndpoints)
	}
}
