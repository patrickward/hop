package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"testing"
	"time"
)

const (
	MailpitSMTPPort = "1025"
	MailpitUIPort   = "8025"
	MailpitImage    = "axllent/mailpit:latest"
	ContainerName   = "mail_test_mailpit"
)

type EmailAddress struct {
	Name    string `json:"Name"`
	Address string `json:"Address"`
}

type MailpitMessage struct {
	ID          string         `json:"ID"`
	MessageID   string         `json:"MessageID"`
	Read        bool           `json:"Read"`
	From        EmailAddress   `json:"From"`
	To          []EmailAddress `json:"To"`
	Subject     string         `json:"Subject"`
	Attachments int            `json:"Attachments"`
	Snippet     string         `json:"Snippet"`
}

type mailpitResponse struct {
	Total         int              `json:"total"`
	Unread        int              `json:"unread"`
	Count         int              `json:"count"`
	MessagesCount int              `json:"messages_count"`
	Start         int              `json:"start"`
	Tags          []string         `json:"tags"`
	Messages      []MailpitMessage `json:"messages"`
}

// SetupMailpit starts a Mailpit container for testing
func SetupMailpit(t *testing.T) func() {
	t.Helper()

	// Check if container is already running
	cmd := exec.Command("docker", "ps", "-q", "-f", fmt.Sprintf("name=%s", ContainerName))
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to check for existing container: %v", err)
	}

	// If container exists, return early
	if len(output) > 0 {
		return func() {} // No cleanup needed for pre-existing container
	}

	// Start Mailpit container
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd = exec.Command("docker", "run", "-d",
		"--name", ContainerName,
		"-p", fmt.Sprintf("%s:1025", MailpitSMTPPort),
		"-p", fmt.Sprintf("%s:8025", MailpitUIPort),
		MailpitImage)

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to start Mailpit container: %v", err)
	}

	// Wait for container to be ready
	time.Sleep(2 * time.Second)

	// Return cleanup function
	return func() {
		cleanupCmd := exec.Command("docker", "rm", "-f", ContainerName)
		if err := cleanupCmd.Run(); err != nil {
			t.Errorf("Failed to cleanup Mailpit container: %v", err)
		}
	}
}

// GetMailpitMessages retrieves messages from Mailpit
func GetMailpitMessages(t *testing.T) []MailpitMessage {
	t.Helper()

	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/api/v1/messages", MailpitUIPort))
	if err != nil {
		t.Fatalf("Failed to get Mailpit messages: %v", err)
	}
	defer resp.Body.Close()

	var response mailpitResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode Mailpit response: %v", err)
	}

	return response.Messages
}

// ClearMailpitMessages clears all messages from Mailpit
func ClearMailpitMessages(t *testing.T) {
	t.Helper()

	req, err := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("http://localhost:%s/api/v1/messages", MailpitUIPort),
		nil)
	if err != nil {
		t.Fatalf("Failed to create delete request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to clear messages: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to clear messages, status: %d", resp.StatusCode)
	}
}

// CheckDockerAvailable verifies if Docker is available
func CheckDockerAvailable(t *testing.T) {
	t.Helper()

	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		t.Skip("Docker not available, skipping mail tests")
	}
}
