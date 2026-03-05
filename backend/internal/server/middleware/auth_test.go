package middleware

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/transport"
)

// mockHeader implements transport.Header
type mockHeader struct {
	values map[string]string
}

func (h *mockHeader) Get(key string) string         { return h.values[key] }
func (h *mockHeader) Set(key, value string)         { h.values[key] = value }
func (h *mockHeader) Add(key, value string)         { h.values[key] = value }
func (h *mockHeader) Keys() []string                { return nil }
func (h *mockHeader) Values(key string) []string    { return []string{h.values[key]} }

// mockTransport implements transport.Transport
type mockTransport struct {
	header *mockHeader
}

func (t *mockTransport) Kind() transport.Kind        { return transport.KindHTTP }
func (t *mockTransport) Endpoint() string            { return "" }
func (t *mockTransport) Operation() string           { return "" }
func (t *mockTransport) RequestHeader() transport.Header { return t.header }
func (t *mockTransport) ReplyHeader() transport.Header   { return t.header }

func ctxWithAuth(authValue string) context.Context {
	tr := &mockTransport{header: &mockHeader{values: map[string]string{
		"Authorization": authValue,
	}}}
	return transport.NewServerContext(context.Background(), tr)
}

func TestAPIKeyAuth(t *testing.T) {
	const key = "test-secret"
	okHandler := func(ctx context.Context, req interface{}) (interface{}, error) { return "ok", nil }
	mw := APIKeyAuth(key)(okHandler)

	tests := []struct {
		name    string
		auth    string
		wantErr bool
	}{
		{"valid bearer", "Bearer test-secret", false},
		{"wrong key", "Bearer wrong", true},
		{"no bearer prefix", "test-secret", true},
		{"empty", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := mw(ctxWithAuth(tc.auth), nil)
			if (err != nil) != tc.wantErr {
				t.Errorf("got err=%v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}
