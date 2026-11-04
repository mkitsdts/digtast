package openai

import (
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestListModels(t *testing.T) {
	httpClient := testHTTPClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/models" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization %q", got)
		}
		return response(http.StatusOK, "application/json", `{"data":[{"id":"b-model"},{"id":"a-model"},{"id":"a-model"}]}`), nil
	})

	names, err := ListModels(context.Background(), Config{APIKey: "test-key", BaseURL: "https://example.test/v1", HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a-model", "b-model"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("got %v, want %v", names, want)
	}
}

func TestListModelsError(t *testing.T) {
	httpClient := testHTTPClient(func(r *http.Request) (*http.Response, error) {
		return response(http.StatusUnauthorized, "application/json", `{"error":{"message":"invalid key"}}`), nil
	})
	if _, err := ListModels(context.Background(), Config{APIKey: "bad", BaseURL: "https://example.test/v1", HTTPClient: httpClient}); err == nil {
		t.Fatal("expected an error for an unauthorized response")
	}
}

func TestProviderDefaults(t *testing.T) {
	if got := ProviderBaseURL("deepseek"); got != "https://api.deepseek.com/v1" {
		t.Fatalf("unexpected deepseek url %q", got)
	}
	if got := ProviderBaseURL("ark"); got != ProviderBaseURL("doubao") {
		t.Fatalf("expected ark to resolve to doubao url, got %q", got)
	}
	if !IsCustomProvider("custom") {
		t.Fatal("custom should require a base url")
	}
}
