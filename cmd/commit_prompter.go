package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/madflow/kommit/internal/logger"
	"github.com/madflow/kommit/internal/workflow"
)

type cliCommitPrompter struct{}

func (cliCommitPrompter) AskForConfirmation() (workflow.CommitAction, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		logger.Printf("Do you want to commit with this message? [y/e/N] ")
		text, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", fmt.Errorf("read confirmation: %w", err)
		}

		text = strings.TrimSpace(strings.ToLower(text))
		switch text {
		case "y", "yes":
			return workflow.CommitActionYes, nil
		case "e", "edit":
			return workflow.CommitActionEdit, nil
		case "", "n", "no":
			return workflow.CommitActionNo, nil
		default:
			logger.Printf("Please enter 'y' for yes, 'e' to edit, or 'N' for no\n")
		}

		if errors.Is(err, io.EOF) {
			return workflow.CommitActionNo, nil
		}
	}
}

func (cliCommitPrompter) EditCommitMessage(message string) (string, error) {
	tempFile, err := os.CreateTemp("", "kommit-*.md")
	if err != nil {
		return "", fmt.Errorf("create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(message); err != nil {
		tempFile.Close()
		return "", fmt.Errorf("write temporary file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("close temporary file: %w", err)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	cmd := exec.Command(editor, tempFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("open editor: %w", err)
	}

	editedMessage, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("read edited message: %w", err)
	}

	return strings.TrimSpace(string(editedMessage)), nil
}
