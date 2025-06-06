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

## Usage

### Configuration

Kommit uses a YAML configuration file to customize its behavior. The configuration file is automatically loaded from one of these locations (in order of priority):

1. `$PWD/.kommit.yaml`
2. `$XDG_CONFIG_HOME/kommit/config.yaml`
3. `$HOME/.config/kommit/config.yaml`
4. `$HOME/.kommit.yaml`

#### Configuration Options

```yaml
# Ollama API configuration
ollama:
  # URL of the Ollama API server (default: http://localhost:11434/api/generate)
  server_url: "http://localhost:11434/api/generate"
  
  # Model to use for generating commit messages (default: "qwen2.5-coder:7b")
  model: "qwen2.5-coder:7b"

# Rules for generating commit messages
# This is a free-form text that guides the AI in generating commit messages
rules: |
  - Start with an emoji that represents the changes (🐛, ✨, 🚀, etc.)
  - Write the first line as if a pirate explaining the changes
  - Include what was changed and why
  - Be creative and have fun with it!
```

### Basic Usage

```bash
# Run kommit in the current git repository
kommit

# Run with a specific config file
kommit --config /path/to/config.yaml
```

### How It Works

1. When you run `kommit`, it will:
   - Check if you're in a git repository
   - Look for uncommitted changes
   - Show a preview of the changes that will be committed
   - Ask for confirmation before committing

2. The tool will automatically:
   - Stage all changes
   - Generate a commit message using the configured Ollama model
   - Show you the generated message
   - Ask for confirmation before creating the commit

### Git Integration

For convenience, you can create a git alias:

```bash
git config --global alias.kommit '!kommit'
```

Then you can use it as:
```bash
git kommit
```

