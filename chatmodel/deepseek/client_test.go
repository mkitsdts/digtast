package deepseek

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"digital-labor/component/chatmodel"
	"digital-labor/schema"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func testHTTPClient(fn roundTripFunc) *http.Client { return &http.Client{Transport: fn} }

func response(status int, contentType, body string) *http.Response {
	return &http.Response{
		StatusCode: status, Status: fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header: http.Header{"Content-Type": []string{contentType}}, Body: io.NopCloser(strings.NewReader(body)),
	}
}

func newTestClient(t *testing.T, model string, httpClient *http.Client) *Client {
	t.Helper()
	client, err := New(Config{APIKey: "test-key", BaseURL: "https://example.test", Model: model, HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestChatParsesReasoningAndCacheUsage(t *testing.T) {
	httpClient := testHTTPClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		return response(http.StatusOK, "application/json", `{"id":"chat-1","model":"deepseek-reasoner","created":1,"choices":[{"index":0,"message":{"role":"assistant","content":"4","reasoning_content":"2+2=4"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15,"prompt_cache_hit_tokens":6,"completion_tokens_details":{"reasoning_tokens":3}}}`), nil
	})
	client := newTestClient(t, Reasoner, httpClient)
	resp, err := client.Chat(context.Background(), &schema.ChatRequest{Messages: []schema.Message{{Role: schema.RoleUser, Content: "2+2?"}}})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Message.Reasoning != "2+2=4" || resp.Message.Content != "4" {
		t.Fatalf("unexpected message %#v", resp.Message)
	}
	if resp.Usage.CachedTokens != 6 || resp.Usage.ReasoningTokens != 3 {
		t.Fatalf("unexpected usage %#v", resp.Usage)
	}
}

func TestChatStreamReasoningDeltas(t *testing.T) {
	httpClient := testHTTPClient(func(r *http.Request) (*http.Response, error) {
		chunks := []string{
			`{"id":"chat-1","model":"deepseek-reasoner","created":1,"choices":[{"index":0,"delta":{"reasoning_content":"先想"}}]}`,
			`{"id":"chat-1","model":"deepseek-reasoner","created":1,"choices":[{"index":0,"delta":{"reasoning_content":"一下"}}]}`,
			`{"id":"chat-1","model":"deepseek-reasoner","created":1,"choices":[{"index":0,"delta":{"content":"答案"},"finish_reason":"stop"}]}`,
			`{"id":"chat-1","model":"deepseek-reasoner","created":1,"choices":[],"usage":{"prompt_tokens":5,"completion_tokens":4,"total_tokens":9,"prompt_cache_hit_tokens":2}}`,
		}
		var body strings.Builder
		for _, chunk := range chunks {
			fmt.Fprintf(&body, "data: %s\n\n", chunk)
		}
		body.WriteString("data: [DONE]\n\n")
		return response(http.StatusOK, "text/event-stream", body.String()), nil
	})
	client := newTestClient(t, Reasoner, httpClient)

	var reasoning, content strings.Builder
	var result *schema.ChatResponse
	for event, err := range client.ChatStream(context.Background(), &schema.ChatRequest{Messages: []schema.Message{{Role: schema.RoleUser, Content: "hi"}}}) {
		if err != nil {
			t.Fatal(err)
		}
		switch event.Type {
		case schema.StreamEventReasoningDelta:
			reasoning.WriteString(event.Delta)
		case schema.StreamEventTextDelta:
			content.WriteString(event.Delta)
		case schema.StreamEventDone:
			result = event.Response
		}
	}
	if result == nil {
		t.Fatal("stream ended without a done event")
	}
	if reasoning.String() != "先想一下" {
		t.Fatalf("unexpected reasoning deltas %q", reasoning.String())
	}
	if content.String() != "答案" || result.Message.Content != "答案" || result.Message.Reasoning != "先想一下" {
		t.Fatalf("unexpected stream result %#v", result.Message)
	}
	if client.GetUsage().TotalTokens != 9 || client.GetUsage().CachedTokens != 2 {
		t.Fatalf("unexpected cumulative usage %#v", client.GetUsage())
	}
}

func TestChatStreamErrorAsEvent(t *testing.T) {
	httpClient := testHTTPClient(func(r *http.Request) (*http.Response, error) {
		return response(http.StatusUnauthorized, "application/json", `{"error":{"message":"invalid key"}}`), nil
	})
	client := newTestClient(t, Chat, httpClient)
	var gotErr error
	for _, err := range client.ChatStream(context.Background(), &schema.ChatRequest{Messages: []schema.Message{{Role: schema.RoleUser, Content: "hi"}}}) {
		if err != nil {
			gotErr = err
		}
	}
	if gotErr == nil || !strings.Contains(gotErr.Error(), "invalid key") {
		t.Fatalf("expected an authorization error, got %v", gotErr)
	}
}

func TestRequestWireFormat(t *testing.T) {
	var payload map[string]any
	httpClient := testHTTPClient(func(r *http.Request) (*http.Response, error) {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		return response(http.StatusOK, "application/json", `{"id":"chat-1","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`), nil
	})
	client := newTestClient(t, Chat, httpClient)
	parallel := true
	_, err := client.Chat(context.Background(), &schema.ChatRequest{
		Messages: []schema.Message{
			{Role: schema.RoleUser, Content: "hi"},
			{Role: schema.RoleAssistant, Content: "hello", Reasoning: "internal thought"},
		},
		MaxOutputTokens:   256,
		ParallelToolCalls: &parallel,
	})
	if err != nil {
		t.Fatal(err)
	}
	if payload["max_tokens"] != float64(256) {
		t.Fatalf("deepseek should use max_tokens, got %v", payload["max_tokens"])
	}
	if _, ok := payload["max_completion_tokens"]; ok {
		t.Fatal("deepseek should not send max_completion_tokens")
	}
	if _, ok := payload["parallel_tool_calls"]; ok {
		t.Fatal("deepseek should not send parallel_tool_calls")
	}
	if strings.Contains(fmt.Sprint(payload["messages"]), "reasoning_content") {
		t.Fatal("reasoning content must not be replayed to deepseek")
	}
	// 未指定模型时使用客户端绑定的默认模型。
	if payload["model"] != Chat {
		t.Fatalf("unexpected model %v", payload["model"])
	}
}

func TestDefaultBaseURLAndRegistration(t *testing.T) {
	client, err := New(Config{APIKey: "k", Model: Chat})
	if err != nil {
		t.Fatal(err)
	}
	if client.base.Model() != Chat {
		t.Fatalf("unexpected model %q", client.base.Model())
	}
	provider, ok := chatmodel.FindProvider("deepseek")
	if !ok {
		t.Fatal("deepseek should be registered")
	}
	if provider.Label == "" || provider.BaseURL != DefaultBaseURL {
		t.Fatalf("unexpected registration %#v", provider)
	}
	if _, err := provider.NewClient(chatmodel.ClientOptions{APIKey: "k"}, Reasoner); err != nil {
		t.Fatalf("build client: %v", err)
	}
}
