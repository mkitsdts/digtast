# SendMessageToSession Test Client

This directory contains a small gRPC client for testing `SendMessageToSession`.
It first calls `StartService` to register an agent, then sends one non-streaming message: `你好`.

## Configure

Edit `config.json` in this directory:

```json
{
  "addr": "127.0.0.1:10086",
  "container_id": "local-container",
  "agent_id": "local-agent",
  "session_id": "test-session",
  "message": "你好",
  "timeout": "2m",
  "provider": "deepseek",
  "api_key": "",
  "base_url": "",
  "model_name": "your-model-name"
}
```

`api_key` can be left empty if `LLM_API_KEY` is set in the environment.

## Run

Start the service from the repository root:

```sh
go run .
```

In another terminal, run the test client from the repository root:

```sh
go run ./test/send_message
```

Or run it from this directory:

```sh
cd test/send_message
go run .
```

## Override Config

Use another config file:

```sh
go run ./test/send_message --config /path/to/config.json
```

Override individual values from the command line:

```sh
go run ./test/send_message --provider deepseek --model your-model-name --key "$LLM_API_KEY"
```

The request always sets `is_stream=false`. The RPC is still defined as server-streaming, so the client reads responses until EOF and prints the combined reply once.
