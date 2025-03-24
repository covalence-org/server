# Netrunner

A firewall for your AI models.

## Overview

Netrunner acts as a middleware layer between your applications and AI service providers (like OpenAI, Anthropic, etc.). It allows you to:

- Register custom aliases for AI models
- Validate requests before they reach the AI provider
- Monitor performance metrics for all requests
- Support streaming responses for real-time AI interactions

This proxy is designed to be simple yet effective, focusing on performance and security without adding unnecessary complexity.

## Installation

### Using Go

```bash
# Clone the repository
git clone https://github.com/ratcht/netrunner.git
cd netrunner

# Build the project
go build -o netrunner

# Run the server
./netrunner
```

### Using Docker

```bash
# Build the Docker image
docker build -t netrunner .

# Run the container
docker run -p 8080:8080 netrunner
```

## Usage

### Registering a Model

```bash
curl -X POST http://localhost:8080/register-model \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-gpt4",
    "model": "gpt-4o",
    "api_url": "https://api.openai.com/v1"
  }'
```

### Listing Registered Models

```bash
curl http://localhost:8080/models
```

### Using with OpenAI Python Client

```python
import openai
import requests

# Register an OpenAI model
response = requests.post(
  "http://localhost:8080/register-model",
  json={
    "name": "my-gpt4",             # Custom name you want to use
    "model": "gpt-4o",             # Actual model name
    "api_url": "https://api.openai.com/v1"  # API URL to forward requests to
  }
)
print(f"Model registration: {response.status_code} - {response.text}")

# Configure the OpenAI client to use your proxy
client = openai.OpenAI(
    api_key="your-api-key",
    base_url="http://localhost:8080/v1"  # Point to your proxy
)

# Make a request using the custom model name
response = client.chat.completions.create(
  model="my-gpt4",  # Use the custom name we registered
  messages=[{"role": "user", "content": "Hello, how are you?"}]
)

print(response.choices[0].message.content)
```

### Using with Streaming

```python
response = client.chat.completions.create(
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

- `POST /register-model`: Register a custom model name
- `GET /models`: List all registered models
- `GET /health`: Health check endpoint
- `ANY /v1/*`: Proxy endpoint that forwards to the appropriate API

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
