# covalence

A firewall for your AI models.

## Overview

covalence acts as a middleware layer between your applications and AI service providers (like OpenAI, Anthropic, etc.). It allows you to:

- Register custom aliases for AI models
- Validate requests before they reach the AI provider
- Monitor performance metrics for all requests
- Support streaming responses for real-time AI interactions

This proxy is designed to be simple yet effective, focusing on performance and security without adding unnecessary complexity.

## Installation

### Using Go

```bash
# Clone the repository
git clone https://github.com/ratcht/covalence.git
cd covalence

# Build the project
go build -o covalence

# Run the server
./covalence
```

### Using Docker

```bash
# Build the Docker image
docker build -t covalence .

# Run the container
docker run -p 8080:8080 covalence
```

## Usage

Head to covalence-org/python-sdk for the official Covalence Python SDK

### Authentication
```python
from covalence import Covalence

# Log into the client
client = Covalence(
  email="me@email.com",
  password="password"
)

print(client.get_user_details())
```


### Registering a Model

```python
client.register_model(
  name="my-gpt",
  model="gpt-4o",
  provider="openai",
  api_key="sk-1234…"
)
```

### Using with OpenAI

```python
from covalence import Covalence
import openai

# 1) Instantiate and authenticate in one line
client = Covalence(
  email="me@email.com",
  password="password"
)

# 2) Register your alias
client.register_model(
  name="my-gpt",
  model="gpt-4o",
  provider="openai",
  api_key="sk-1234…"
)

# 3) Pump the token into your existing OpenAI client
openai_client = openai.OpenAI(
  api_key=client.token(),
  base_url=client.url(),
)

# 4) Make a chat call
resp = openai_client.chat.completions.create(
  model="my-gpt",
  messages=[{"role":"user","content":"Hello, how are you?"}],
)
print(resp.choices[0].message.content)
```

### Using with Streaming

```python
response = openai_client.chat.completions.create(
  model="my-gpt4",
  messages=[{"role": "user", "content": "Write a story about a robot."}],
  stream=True  # Enable streaming
)

for chunk in response:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

## Features

- **Model Aliasing**: Register custom model names that map to actual provider models
- **Request Validation**: Validate request parameters before forwarding to prevent errors
- **Performance Metrics**: Track and log detailed metrics for each request
- **Streaming Support**: Properly handle streaming API responses
- **Simple API**: Easy-to-use REST API for management and proxying
- **High Performance**: Optimized for low latency and high throughput

## Configuration

The server runs on port 8080 by default. You can modify the code to change this or add environment variable support.

## API Endpoints

- Navigate to /swagger/docs for available endpoints

## Performance Metrics

The proxy logs detailed metrics for each request in JSON format, including:

- Total processing time
- Model lookup time
- Request body processing time
- Upstream service latency
- Status code
- Model information
- Streaming status

Example log:

```
REQUEST_METRICS: {"timestamp":"2025-03-22T12:34:56Z","custom_model":"my-gpt4","actual_model":"gpt-4o","status":200,"lookup_ms":2,"body_process_ms":5,"upstream_ms":1250,"total_ms":1258,"streaming":true,"path":"/chat/completions"}
```

## Security Considerations

This proxy includes:

- Input validation for model parameters
- Safe header forwarding
- Timeout protection
- Request body validation

Additional security measures like rate limiting, authentication, and TLS can be added as needed.

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
