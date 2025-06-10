# Kommit

> Git commits for the rest of us

## Installation

### macOS

```bash
brew install --no-quarantine madflow/kommit/kommit
```

### Docker

```bash
# Build the image
docker build -t madflow/kommit .

# Basic usage with Git configuration
docker run -it --rm \
  -v $PWD:/workdir \
  -e GIT_USER_NAME="Your Name" \
  -e GIT_USER_EMAIL="your.email@example.com" \
  madflow/kommit [command] [args...]

# With Ollama running on host (Linux/macOS)
docker run -it --rm \
  -v $PWD:/workdir \
  --network=host \
  -e OLLAMA_HOST=host.docker.internal \
  -e GIT_USER_NAME="Your Name" \
  -e GIT_USER_EMAIL="your.email@example.com" \
  madflow/kommit [command] [args...]

# With explicit Ollama host (Windows/Linux/macOS)
docker run -it --rm \
  -v $PWD:/workdir \
  -e OLLAMA_HOST=host.docker.internal \
  -e GIT_USER_NAME="Your Name" \
  -e GIT_USER_EMAIL="your.email@example.com" \
  --add-host=host.docker.internal:host-gateway \
  madflow/kommit [command] [args...]

# Using automatic Git config from host
docker run -it --rm \
  -v $PWD:/workdir \
  --network=host \
  -e GIT_USER_NAME="$(git config --get user.name)" \
  -e GIT_USER_EMAIL="$(git config --get user.email)" \
  -e OLLAMA_HOST=host.docker.internal \
  --add-host=host.docker.internal:host-gateway \
  madflow/kommit

# Using environment file with automatic Git config
echo "GIT_USER_NAME=$(git config --get user.name)" > .kommit.env
echo "GIT_USER_EMAIL=$(git config --get user.email)" >> .kommit.env
echo "OLLAMA_HOST=host.docker.internal" >> .kommit.env

docker run -it --rm \
  -v $PWD:/workdir \
  --env-file .kommit.env \
  --add-host=host.docker.internal:host-gateway \
  madflow/kommit
```

#### Git Configuration in Docker

When running Kommit in a Docker container, you can provide Git user configuration in several ways:

1. **Automatic (Recommended)**:

   ```bash
   -e GIT_USER_NAME="$(git config --get user.name)" \
   -e GIT_USER_EMAIL="$(git config --get user.email)"
   ```

2. **Manual**:

   ```bash
   -e GIT_USER_NAME="Your Name" \
   -e GIT_USER_EMAIL="your.email@example.com"
   ```

3. **Via .env file**:
   ```bash
   echo "GIT_USER_NAME=$(git config --get user.name)" > .kommit.env
   echo "GIT_USER_EMAIL=$(git config --get user.email)" >> .kommit.env
   docker run --env-file .kommit.env ...
   ```

If no Git configuration is provided, it will use these defaults:

- `GIT_USER_NAME`: "Kommit User"
- `GIT_USER_EMAIL`: "kommit@example.com"

## Usage

### Configuration

Kommit uses a YAML configuration file to customize its behavior. The configuration file is automatically loaded from one of these locations (in order of priority):

1. `$PWD/.kommit.yaml`
2. `$GIT_DIR/.konfig.yaml` (if inside a git repository)
3. `$XDG_CONFIG_HOME/kommit/config.yaml`
4. `$HOME/.config/kommit/config.yaml`
5. `$HOME/.kommit.yaml`

> **Note**: The `$GIT_DIR` location is particularly useful for repository-specific configurations that should be shared with all contributors.

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

# YOLO mode: Automatically stage, commit, and push changes (no confirmation)
kommit --yolo
# or use the short flag
kommit -y
```

### How It Works

When you run `kommit`, it will:

- Check if you're in a git repository
- Check if there is a valid config file in one of the supported locations
- Use the defaults if no config file is found
- Generate a commit message using the configured Ollama model
- Show a preview of the changes that will be committed
- Ask for confirmation before committing

### Git Integration

For convenience, you can create a git alias:

```bash
git config --global alias.kommit '!kommit'
```

Then you can use it as:

```bash
git kommit
```
