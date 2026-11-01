package client

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
)

// NNTPClient represents a client for connecting to NNTP servers
type NNTPClient struct {
	// Connection details
	server   *models.NNTPServer
	conn     net.Conn
	timeout  time.Duration
	maxRetries int
}

// NewNNTPClient creates a new NNTP client for the given server
func NewNNTPClient(server *models.NNTPServer, timeout time.Duration, maxRetries int) *NNTPClient {
	return &NNTPClient{
		server:   server,
		timeout:  timeout,
		maxRetries: maxRetries,
	}
}

// Connect establishes a connection to the NNTP server
func (c *NNTPClient) Connect(ctx context.Context) error {
	// Create connection string
	address := fmt.Sprintf("%s:%d", c.server.Hostname, c.server.Port)

	// Use context with timeout for connection
	connCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	var conn net.Conn
	var err error

	// Attempt connection with retries
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Second * time.Duration(attempt)) // Exponential backoff
		}

		conn, err = net.DialTimeout("tcp", address, c.timeout)
		if err == nil {
			c.conn = conn
			return nil
		}
	}

	return fmt.Errorf("failed to connect to NNTP server %s after %d attempts", c.server.Hostname, c.maxRetries+1)
}

// Disconnect closes the connection to the NNTP server
func (c *NNTPClient) Disconnect() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected checks if the client is currently connected
func (c *NNTPClient) IsConnected() bool {
	return c.conn != nil
}

// SendCommand sends a command to the NNTP server and returns the response
func (c *NNTPClient) SendCommand(ctx context.Context, command string, args ...string) (int, string, error) {
	if c.conn == nil {
		return 0, "", fmt.Errorf("not connected to NNTP server")
	}

	// Format the command
	cmd := command
	if len(args) > 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}

	// Send command
	_, err := fmt.Fprintf(c.conn, "%s\r\n", cmd)
	if err != nil {
		return 0, "", fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	buffer := make([]byte, 1024)
	n, err := c.conn.Read(buffer)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read response: %w", err)
	}

	response := string(buffer[:n])
	lines := strings.Split(strings.TrimSpace(response), "\r\n")
	if len(lines) > 0 {
		responseLine := lines[0]
		parts := strings.SplitN(responseLine, " ", 2)
		if len(parts) >= 1 {
			statusCode, err := strconv.Atoi(parts[0])
			if err != nil {
				return 0, responseLine, fmt.Errorf("invalid status code in response: %w", err)
			}
			if len(parts) > 1 {
				return statusCode, parts[1], nil
			}
			return statusCode, "", nil
		}
	}

	return 0, response, fmt.Errorf("unexpected response format: %s", response)
}

// Authenticate authenticates to the NNTP server with username and password
func (c *NNTPClient) Authenticate(ctx context.Context, username, password string) error {
	if c.conn == nil {
		return fmt.Errorf("not connected to NNTP server")
	}

	// Send authentication command
	status, response, err := c.SendCommand(ctx, "AUTHINFO", "USER", username)
	if err != nil {
		return fmt.Errorf("failed to send AUTHINFO USER: %w", err)
	}

	if status >= 400 {
		return fmt.Errorf("authentication failed (USER): %s", response)
	}

	// Send password command
	status, response, err = c.SendCommand(ctx, "AUTHINFO", "PASS", password)
	if err != nil {
		return fmt.Errorf("failed to send AUTHINFO PASS: %w", err)
	}

	if status >= 400 {
		return fmt.Errorf("authentication failed (PASS): %s", response)
	}

	return nil
}