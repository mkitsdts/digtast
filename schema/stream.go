package schema

import "iter"

type StreamEventType string

const (
	StreamEventReasoningDelta StreamEventType = "reasoning_delta"
	StreamEventTextDelta      StreamEventType = "text_delta"
	StreamEventToolCallDelta  StreamEventType = "tool_call_delta"
	StreamEventUsage          StreamEventType = "usage"
	StreamEventDone           StreamEventType = "done"
)

type ToolCallDelta struct {
	Index     int
	ID        string
	Type      string
	Name      string
	Arguments string
}

// StreamEvent 是流式对话产出的事件；Done 事件携带装配完成的 ChatResponse。
type StreamEvent struct {
	Type     StreamEventType
	Delta    string
	ToolCall *ToolCallDelta
	Usage    *Usage
	Response *ChatResponse
}

// Stream 是流式对话返回的事件迭代器：正常情况下依次产出 (event, nil)，
// 以 Done 事件收尾；失败时产出一次 (nil, err) 并终止。调用方提前退出迭代
// 时，迭代器内部会释放底层连接。
type Stream iter.Seq2[*StreamEvent, error]

// SliceStream 把固定事件封装成 Stream，主要用于测试。
func SliceStream(events ...*StreamEvent) Stream {
	return func(yield func(*StreamEvent, error) bool) {
		for _, event := range events {
			if !yield(event, nil) {
				return
			}
		}
	}
}

// ErrorStream 返回只产出一个错误的 Stream，主要用于测试。
func ErrorStream(err error) Stream {
	return func(yield func(*StreamEvent, error) bool) {
		yield(nil, err)
	}
}
