package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"digital-labor/pkg/llm"
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

func TestClientChatAndUsage(t *testing.T) {
	httpClient := testHTTPClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization %q", got)
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request["model"] != "test-model" {
			t.Fatalf("unexpected model %v", request["model"])
		}
		return response(http.StatusOK, "application/json", `{"id":"chat-1","model":"test-model","created":1,"choices":[{"index":0,"message":{"role":"assistant","content":"hello"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":2,"total_tokens":5}}`), nil
	})

	client, err := New(Config{APIKey: "test-key", BaseURL: "https://example.test/v1", Model: "test-model", HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Chat(context.Background(), &llm.ChatRequest{Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}}})
	if err != nil {
		t.Fatal(err)
	}
	if response.Message.Content != "hello" || response.Usage.TotalTokens != 5 {
		t.Fatalf("unexpected response %#v", response)
	}
	if usage := client.GetUsage(); usage.InputTokens != 3 || usage.OutputTokens != 2 || usage.TotalTokens != 5 {
		t.Fatalf("unexpected cumulative usage %#v", usage)
	}
}

func TestClientChatStreamToolCall(t *testing.T) {
	httpClient := testHTTPClient(func(r *http.Request) (*http.Response, error) {
		chunks := []string{
			`{"id":"chat-1","model":"test-model","created":1,"choices":[{"index":0,"delta":{"role":"assistant","content":"checking "},"finish_reason":""}]}`,
			`{"id":"chat-1","model":"test-model","created":1,"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call-1","type":"function","function":{"name":"weather","arguments":"{\"city\":"}}]},"finish_reason":""}]}`,
			`{"id":"chat-1","model":"test-model","created":1,"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"gz\"}"}}]},"finish_reason":"tool_calls"}]}`,
			`{"id":"chat-1","model":"test-model","created":1,"choices":[],"usage":{"prompt_tokens":4,"completion_tokens":3,"total_tokens":7}}`,
		}
		var body strings.Builder
		for _, chunk := range chunks {
			fmt.Fprintf(&body, "data: %s\n\n", chunk)
		}
		body.WriteString("data: [DONE]\n\n")
		return response(http.StatusOK, "text/event-stream", body.String()), nil
	})
	client, err := New(Config{APIKey: "test-key", BaseURL: "https://example.test/v1", Model: "test-model", HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	stream, err := client.ChatStream(context.Background(), &llm.ChatRequest{Messages: []llm.Message{{Role: llm.RoleUser, Content: "weather"}}})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	var response *llm.ChatResponse
	for response == nil {
		event, err := stream.Recv()
		if err != nil {
			t.Fatal(err)
		}
		if event.Type == llm.StreamEventDone {
			response = event.Response
		}
	}
	if len(response.Message.ToolCalls) != 1 {
		t.Fatalf("unexpected tool calls %#v", response.Message.ToolCalls)
	}
	call := response.Message.ToolCalls[0]
	if call.Function.Name != "weather" || !strings.Contains(string(call.Function.Arguments), `"city":"gz"`) {
		t.Fatalf("unexpected tool call %#v", call)
	}
	if client.GetUsage().TotalTokens != 7 {
		t.Fatalf("stream usage was not accumulated")
	}
}

func TestCheckStatusFallsBackToModelsList(t *testing.T) {
	httpClient := testHTTPClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/v1/models/test-model" {
			return response(http.StatusNotFound, "application/json", `{"error":{"message":"not found"}}`), nil
		}
		if r.URL.Path != "/v1/models" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		return response(http.StatusOK, "application/json", `{"object":"list","data":[]}`), nil
	})
	client, _ := New(Config{APIKey: "test-key", BaseURL: "https://example.test/v1", Model: "test-model", HTTPClient: httpClient})
	status, err := client.CheckStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !status.Available || status.Model != "test-model" {
		t.Fatalf("unexpected status %#v", status)
	}
}
