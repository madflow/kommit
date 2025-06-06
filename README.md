# Kommit

> Git commits for the rest of us

## Installation

### Docker

```bash
# Build the image
docker build -t madflow/kommit .

# Basic usage
docker run -it --rm -v $PWD:/workdir madflow/kommit [command] [args...]

# With Ollama running on host (Linux/macOS)
docker run -it --rm \
  -v $PWD:/workdir \
  --network=host \
  -e OLLAMA_HOST=host.docker.internal \
  madflow/kommit [command] [args...]

# With explicit Ollama host (Windows/Linux/macOS)
docker run -it --rm \
  -v $PWD:/workdir \
  -e OLLAMA_HOST=host.docker.internal \
  --add-host=host.docker.internal:host-gateway \
  madflow/kommit [command] [args...]
```

