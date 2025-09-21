# Kommit

> Git commits for the rest of us

`kommit` is a command-line tool that uses the power of AI to generate git commit messages for you. By analyzing the changes in your code, it creates descriptive and relevant commit messages, saving you time and effort. It is designed to work with Ollama to generate commit messages and is highly configurable via a configuration file. You can either accept the generated message as is, edit it to your liking, or use the "YOLO" mode to automatically commit and push your changes without confirmation.

## Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
  - [macOS](#macos)
  - [Docker](#docker)
    - [Git Configuration in Docker](#git-configuration-in-docker)
- [Configuration](#configuration)
  - [Configuration Options](#configuration-options)
- [Usage](#usage)
  - [Basic Usage](#basic-usage)
  - [How It Works](#how-it-works)
  - [Commit Message Editing](#commit-message-editing)
  - [Git Integration](#git-integration)

## Requirements

- [Ollama](https://ollama.ai/) - Required for generating commit messages. Make sure Ollama is installed and running before using Kommit.
  - Download and install from: https://ollama.ai/download
  - Start the Ollama server: `ollama serve`

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

### Basic Usage

```bash
# Run kommit in the current git repository
kommit

# Stage all changes before committing
kommit --add
# or use the short flag
kommit -a

# Run with a specific config file
kommit --config /path/to/config.yaml

# YOLO mode: Automatically stage, commit, and push changes (no confirmation)
kommit --yolo
# or use the short flag
kommit -y

# Create a pull request after committing (requires GitHub repository and gh CLI)
kommit --pr

# Combine options: YOLO mode + create pull request
kommit --yolo --pr
```

### Options

#### `branch`

This subcommand generates a new Git branch name based on your current staged changes and creates the branch. It uses the configured Ollama model to analyze the diff and suggest a relevant branch name, ensuring it's no more than 50 characters long. This is useful for quickly creating feature or bugfix branches that reflect the work you're about to commit.

Usage:
```bash
kommit branch
```

#### `--add` or `-a`

This option automatically stages all changes in your Git repository before generating a commit message. It's equivalent to running `git add .` before `kommit`. Use this when you want to quickly commit all modified and new files without manually staging them.

#### `--yolo` or `-y`

This option automatically stages all changes, commits with the generated message, and pushes to the remote repository without any confirmation prompts. It's a "fire-and-forget" mode for when you trust the process completely.

#### `--pr`

This option creates a pull request after committing changes. The behavior depends on your current branch:

- **On the main branch**: When you're on the origin's main branch (automatically detected), the --pr flag has no effect and does nothing
- **On a feature branch**: Creates a pull request against the origin main branch using both local and remote changes for AI generation

The AI analyzes both:
- **Local changes**: Your currently staged changes (what will be committed)
- **Remote changes**: All changes in your branch since it diverged from the origin main branch

This option requires:
- A GitHub repository
- GitHub CLI (`gh`) installed and authenticated  
- Remote tracking set up (`git remote set-head origin -a` if needed)

When used with `--yolo`, it will automatically commit, push, and create a pull request in one command. The AI-generated pull request title and description follow configurable rules and can be customized in your configuration file.

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

# Rules for generating pull request descriptions
pr_rules: |
  - Create a concise summary for a pull request in the format "## Summary" followed by 1-3 bullet points
  - Each bullet point should highlight a key change or improvement made in this pull request
  - Focus on the value and impact of the changes, not just what files were modified
  - Use simple, direct language that explains what was accomplished

# Rules for generating pull request titles  
pr_title_rules: |
  - Create a concise and descriptive pull request title
  - Maximum length: 50 characters (aim for under 50 characters)
  - Use imperative mood ("Add feature" not "Added feature" or "Adds feature")
  - Start with a verb when possible (Add, Fix, Update, Remove, etc.)
  - Be specific about what was changed or accomplished
```

### Basic Usage

```bash
# Run kommit in the current git repository
kommit

# Stage all changes before committing
kommit --add
# or use the short flag
kommit -a

# Run with a specific config file
kommit --config /path/to/config.yaml

# YOLO mode: Automatically stage, commit, and push changes (no confirmation)
kommit --yolo
# or use the short flag
kommit -y
```

### Options

#### `--add` or `-a`

This option automatically stages all changes in your Git repository before generating a commit message. It's equivalent to running `git add .` before `kommit`. Use this when you want to quickly commit all modified and new files without manually staging them.

#### `--yolo` or `-y`

This option automatically stages all changes, commits with the generated message, and pushes to the remote repository without any confirmation prompts. It's a "fire-and-forget" mode for when you trust the process completely.

#### `--pr`

This option creates a pull request after committing changes. The behavior depends on your current branch:

- **On the main branch**: When you're on the origin's main branch (automatically detected), the --pr flag has no effect and does nothing
- **On a feature branch**: Creates a pull request against the origin main branch using both local and remote changes for AI generation

The AI analyzes both:
- **Local changes**: Your currently staged changes (what will be committed)
- **Remote changes**: All changes in your branch since it diverged from the origin main branch

This option requires:
- A GitHub repository
- GitHub CLI (`gh`) installed and authenticated  
- Remote tracking set up (`git remote set-head origin -a` if needed)

When used with `--yolo`, it will automatically commit, push, and create a pull request in one command. The AI-generated pull request title and description follow configurable rules and can be customized in your configuration file.

### How It Works

When you run `kommit`, it will:

- Check if you're in a git repository
- Check if there is a valid config file in one of the supported locations
- Use the defaults if no config file is found
- Generate a commit message using the configured Ollama model
- Show a preview of the changes that will be committed
- Ask for confirmation before committing with the following options:
  - `y` or `yes`: Accept the generated message and commit
  - `e` or `edit`: Open your default editor to modify the commit message
  - `n` or `no` (or press Enter): Cancel the commit

### Commit Message Editing

When you choose to edit the commit message (`e` or `edit`):
1. Your default text editor will open with the generated commit message
   - The editor is determined by the `$EDITOR` environment variable, or defaults to `vi`
2. Make your changes to the commit message and save the file
3. After closing the editor, you'll see the updated message and be prompted again
4. You can continue editing as many times as needed until you're satisfied
5. Choose `y` to commit with the current message or `n` to cancel

### Git Integration

For convenience, you can create a git alias:

```bash
git config --global alias.kommit '!kommit'
```

Then you can use it as:

```bash
git kommit
```
