package client

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
)

// NNTPServerInfo represents a newsgroup with its article ranges
type NewsgroupInfo struct {
	Name           string
	FirstArticle   int64
	LastArticle    int64
	Flags          string
	Description    string
}

// ArticleOverview represents an article overview from XOVER/XHDR
type ArticleOverview struct {
	Number      int
	Subject     string
	From        string
	Date        string
	MessageID   string
	Lines       int
	Size        int
	Raw         string // raw XOVER response line
}

// ArticleBody represents a full article with headers and body
type ArticleBody struct {
	Headers map[string]string
	Body    string
}

// ArticleHeaders represents the parsed headers of an article
type ArticleHeaders struct {
	Subject   string
	From      string
	Date      string
	MessageID string
	Lines     int
	Size      int
}

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
	// Create connection string (handle IPv6 with brackets)
	address := net.JoinHostPort(c.server.Hostname, strconv.Itoa(c.server.Port))

	// Use context with timeout for connection
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
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

// Disconnect closes the connection to the NNTP server. Safe to call multiple times.
func (c *NNTPClient) Disconnect() error {
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

// IsConnected checks if the client is currently connected
func (c *NNTPClient) IsConnected() bool {
	return c.conn != nil
}

// SendCommand sends a command to the NNTP server and handles multi-line responses.
// NNTP multi-line responses end with a line containing only a period (".").
// The period line is stripped from the output.
// Returns the status code, the response text from the status line (after the code),
// and the full body (for multi-line responses) or empty string (for single-line).
func (c *NNTPClient) SendCommand(ctx context.Context, command string, args ...string) (int, string, string, error) {
	if c.conn == nil {
		return 0, "", "", fmt.Errorf("not connected to NNTP server")
	}

	// Set read deadline for the response
	if c.timeout > 0 {
		if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
			return 0, "", "", fmt.Errorf("failed to set read deadline: %w", err)
		}
	}

	// Format the command
	cmd := command
	if len(args) > 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}

	// Send command
	_, err := fmt.Fprintf(c.conn, "%s\r\n", cmd)
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to send command: %w", err)
	}

	// Read the first line to get the status code
	reader := bufio.NewReader(c.conn)
	firstLine, err := reader.ReadString('\n')
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to read response status line: %w", err)
	}

	// Parse status code: "NNNN response text"
	statusCode, responseText, err := parseStatusLine(firstLine)
	if err != nil {
		return 0, "", "", err
	}

	// If status is 2xx or 3xx and the response isn't a single-line response,
	// read the multi-line body until we see a line starting with "."
	if statusCode >= 200 && statusCode < 400 {
		body, err := readMultiLineResponse(reader)
		if err != nil {
			return statusCode, responseText, "", fmt.Errorf("failed to read multi-line response: %w", err)
		}
		return statusCode, responseText, body, nil
	}

	// For 4xx and 5xx errors, return the status line text (may contain multi-line error)
	// Check if there are continuation lines starting with "-"
	if statusCode >= 400 {
		body, err := readMultiLineResponse(reader)
		if err != nil {
			return statusCode, responseText, "", fmt.Errorf("failed to read multi-line error response: %w", err)
		}
		return statusCode, responseText, body, nil
	}

	return statusCode, responseText, "", nil
}

// parseStatusLine parses an NNTP status line like "215 Welcome to NNTP"
func parseStatusLine(line string) (int, string, error) {
	line = strings.TrimSpace(line)
	if len(line) < 4 {
		return 0, "", fmt.Errorf("invalid status line: %s", line)
	}

	statusCode, err := strconv.Atoi(line[:3])
	if err != nil {
		return 0, "", fmt.Errorf("invalid status code in response: %s", line)
	}

	responseText := ""
	if len(line) > 4 {
		responseText = line[4:]
	}

	return statusCode, responseText, nil
}

// readMultiLineResponse reads lines until one starts with "." (the NNTP end marker).
// The "." line is stripped from the output.
func readMultiLineResponse(reader *bufio.Reader) (string, error) {
	var lines []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if len(lines) > 0 {
				// We have data but hit an error; return what we got
				return strings.Join(lines, "\n"), nil
			}
			return "", fmt.Errorf("failed to read multi-line response: %w", err)
		}

		// Remove trailing newline
		line = strings.TrimRight(line, "\r\n")

		// Check for the end marker: a line containing only "."
		if line == "." {
			break
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n"), nil
}

// Authenticate authenticates to the NNTP server with username and password
func (c *NNTPClient) Authenticate(ctx context.Context, username, password string) error {
	if c.conn == nil {
		return fmt.Errorf("not connected to NNTP server")
	}

	// Send authentication command
	status, responseText, _, err := c.SendCommand(ctx, "AUTHINFO", "USER", username)
	if err != nil {
		return fmt.Errorf("failed to send AUTHINFO USER: %w", err)
	}

	if status >= 400 {
		return fmt.Errorf("authentication failed (USER): %s", responseText)
	}

	// Send password command
	status, responseText, _, err = c.SendCommand(ctx, "AUTHINFO", "PASS", password)
	if err != nil {
		return fmt.Errorf("failed to send AUTHINFO PASS: %w", err)
	}

	if status >= 400 {
		return fmt.Errorf("authentication failed (PASS): %s", responseText)
	}

	return nil
}

// List returns all newsgroups with their article ranges.
// The LIST response format is: "name high low flags" or "name high low"
func (c *NNTPClient) List(ctx context.Context) ([]NewsgroupInfo, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to NNTP server")
	}

	status, _, body, err := c.SendCommand(ctx, "LIST")
	if err != nil {
		return nil, fmt.Errorf("LIST command failed: %w", err)
	}

	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("LIST command returned error status %d: %s", status, body)
	}

	return parseNewsgroups(body)
}

// ListActive returns newsgroups with active file format:
// "name high low flags description"
func (c *NNTPClient) ListActive(ctx context.Context) ([]NewsgroupInfo, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to NNTP server")
	}

	status, _, body, err := c.SendCommand(ctx, "LIST ACTIVE")
	if err != nil {
		return nil, fmt.Errorf("LIST ACTIVE command failed: %w", err)
	}

	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("LIST ACTIVE command returned error status %d: %s", status, body)
	}

	return parseNewsgroupsActive(body)
}

// parseNewsgroups parses a LIST response body into NewsgroupInfo slices
func parseNewsgroups(body string) ([]NewsgroupInfo, error) {
	var groups []NewsgroupInfo
	lines := strings.Split(body, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		group := NewsgroupInfo{
			Name: parts[0],
		}

		var err error
		group.LastArticle, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}
		group.FirstArticle, err = strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			continue
		}

		if len(parts) > 3 {
			group.Flags = parts[3]
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// parseNewsgroupsActive parses a LIST ACTIVE response body into NewsgroupInfo slices.
// Format: "name high low flags description"
func parseNewsgroupsActive(body string) ([]NewsgroupInfo, error) {
	var groups []NewsgroupInfo
	lines := strings.Split(body, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		group := NewsgroupInfo{
			Name: parts[0],
		}

		var err error
		group.LastArticle, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}
		group.FirstArticle, err = strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			continue
		}

		if len(parts) > 3 {
			group.Flags = parts[3]
		}

		if len(parts) > 4 {
			group.Description = strings.Join(parts[4:], " ")
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// Xover fetches article overviews for a range of articles.
// XOVER format: "number Subject\tFrom\tDate\tMessage-ID\tLines\tSize"
func (c *NNTPClient) Xover(ctx context.Context, start, end int) ([]ArticleOverview, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to NNTP server")
	}

	status, _, body, err := c.SendCommand(ctx, "XOVER", strconv.Itoa(start), strconv.Itoa(end))
	if err != nil {
		return nil, fmt.Errorf("XOVER command failed: %w", err)
	}

	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("XOVER command returned error status %d: %s", status, body)
	}

	return parseXover(body)
}

// XoverByMessageIDs fetches overviews for specific message IDs using XHDR.
func (c *NNTPClient) XoverByMessageIDs(ctx context.Context, messageIDs []string) ([]ArticleOverview, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to NNTP server")
	}

	if len(messageIDs) == 0 {
		return []ArticleOverview{}, nil
	}

	overviews := make(map[string]*ArticleOverview)

	// Fetch subject for each message ID
	for _, msgID := range messageIDs {
		status, _, body, err := c.SendCommand(ctx, "XHDR", "subject", msgID)
		if err != nil {
			continue
		}
		if status < 200 || status >= 300 {
			continue
		}
		// XHDR subject response: "number Subject line"
		if body != "" {
			parts := strings.SplitN(body, "\t", 2)
			if len(parts) == 2 {
				num, err := strconv.Atoi(parts[0])
				if err != nil {
					continue
				}
				overviews[msgID] = &ArticleOverview{
					Number:  num,
					Subject: parts[1],
				}
			}
		}
	}

	// Fetch from for each message ID
	for _, msgID := range messageIDs {
		if _, ok := overviews[msgID]; !ok {
			continue
		}

		status, _, body, err := c.SendCommand(ctx, "XHDR", "from", msgID)
		if err != nil {
			continue
		}
		if status < 200 || status >= 300 {
			continue
		}
		if body != "" {
			parts := strings.SplitN(body, "\t", 2)
			if len(parts) == 2 {
				overviews[msgID].From = parts[1]
			}
		}
	}

	// Fetch date for each message ID
	for _, msgID := range messageIDs {
		if _, ok := overviews[msgID]; !ok {
			continue
		}

		status, _, body, err := c.SendCommand(ctx, "XHDR", "date", msgID)
		if err != nil {
			continue
		}
		if status < 200 || status >= 300 {
			continue
		}
		if body != "" {
			parts := strings.SplitN(body, "\t", 2)
			if len(parts) == 2 {
				overviews[msgID].Date = parts[1]
			}
		}
	}

	// Fetch message-id for each message ID
	for _, msgID := range messageIDs {
		if _, ok := overviews[msgID]; !ok {
			continue
		}

		status, _, body, err := c.SendCommand(ctx, "XHDR", "message-id", msgID)
		if err != nil {
			continue
		}
		if status < 200 || status >= 300 {
			continue
		}
		if body != "" {
			parts := strings.SplitN(body, "\t", 2)
			if len(parts) == 2 {
				overviews[msgID].MessageID = parts[1]
			}
		}
	}

	// Fetch lines and size for each message ID
	for _, msgID := range messageIDs {
		if _, ok := overviews[msgID]; !ok {
			continue
		}

		status, _, body, err := c.SendCommand(ctx, "XHDR", "lines", msgID)
		if err != nil {
			continue
		}
		if status < 200 || status >= 300 {
			continue
		}
		if body != "" {
			parts := strings.SplitN(body, "\t", 2)
			if len(parts) == 2 {
				lines, err := strconv.Atoi(parts[1])
				if err == nil {
					overviews[msgID].Lines = lines
				}
			}
		}

		status, _, body, err = c.SendCommand(ctx, "XHDR", "size", msgID)
		if err != nil {
			continue
		}
		if status < 200 || status >= 300 {
			continue
		}
		if body != "" {
			parts := strings.SplitN(body, "\t", 2)
			if len(parts) == 2 {
				size, err := strconv.Atoi(parts[1])
				if err == nil {
					overviews[msgID].Size = size
				}
			}
		}
	}

	// Convert map to slice
	var result []ArticleOverview
	for _, overview := range overviews {
		result = append(result, *overview)
	}

	return result, nil
}

// parseXover parses an XOVER response body into ArticleOverview slices.
// Format: "number Subject\tFrom\tDate\tMessage-ID\tLines\tSize"
func parseXover(body string) ([]ArticleOverview, error) {
	var overviews []ArticleOverview
	lines := strings.Split(body, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 6 {
			continue
		}

		overview := ArticleOverview{
			Raw: line,
		}

		var err error
		overview.Number, err = strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		overview.Subject = parts[1]
		overview.From = parts[2]
		overview.Date = parts[3]
		overview.MessageID = parts[4]

		overview.Lines, err = strconv.Atoi(parts[5])
		if err != nil {
			overview.Lines = 0
		}

		if len(parts) > 6 {
			overview.Size, err = strconv.Atoi(parts[6])
			if err != nil {
				overview.Size = 0
			}
		}

		overviews = append(overviews, overview)
	}

	return overviews, nil
}

// ArticleByNumber fetches an article by its number.
// Returns headers (parsed) and the full body.
func (c *NNTPClient) ArticleByNumber(ctx context.Context, number int) (*ArticleBody, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to NNTP server")
	}

	status, _, body, err := c.SendCommand(ctx, "ARTICLE", strconv.Itoa(number))
	if err != nil {
		return nil, fmt.Errorf("ARTICLE command failed: %w", err)
	}

	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("ARTICLE command returned error status %d", status)
	}

	return parseArticle(body)
}

// ArticleByMessageID fetches an article by its message ID.
func (c *NNTPClient) ArticleByMessageID(ctx context.Context, messageID string) (*ArticleBody, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to NNTP server")
	}

	status, _, body, err := c.SendCommand(ctx, "ARTICLE", "<"+messageID+">")
	if err != nil {
		return nil, fmt.Errorf("ARTICLE command failed: %w", err)
	}

	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("ARTICLE command returned error status %d", status)
	}

	return parseArticle(body)
}

// parseArticle parses an ARTICLE response into headers and body.
// The response format is: headers followed by a blank line, then the body.
func parseArticle(body string) (*ArticleBody, error) {
	article := &ArticleBody{
		Headers: make(map[string]string),
	}

	// Split at the blank line that separates headers from body
	parts := strings.SplitN(body, "\n\n", 2)
	headerText := parts[0]

	// Parse headers
	for _, line := range strings.Split(headerText, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Handle folded headers (continuation lines start with whitespace)
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			// Append to the last header
			// Find the last key we added
			continue
		}

		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}

		key := strings.TrimSpace(line[:colonIdx])
		value := strings.TrimSpace(line[colonIdx+1:])

		// NNTP headers are case-insensitive; normalize to canonical form
		key = strings.ToLower(key)
		article.Headers[key] = value
	}

	if len(parts) > 1 {
		article.Body = parts[1]
	}

	return article, nil
}

// Head fetches only the headers of an article by number.
func (c *NNTPClient) Head(ctx context.Context, number int) (*ArticleHeaders, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to NNTP server")
	}

	status, _, body, err := c.SendCommand(ctx, "HEAD", strconv.Itoa(number))
	if err != nil {
		return nil, fmt.Errorf("HEAD command failed: %w", err)
	}

	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("HEAD command returned error status %d", status)
	}

	return parseHeaders(body)
}

// parseHeaders parses header text into an ArticleHeaders struct
func parseHeaders(body string) (*ArticleHeaders, error) {
	headers := &ArticleHeaders{}

	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(line[:colonIdx]))
		value := strings.TrimSpace(line[colonIdx+1:])

		switch key {
		case "subject":
			headers.Subject = value
		case "from":
			headers.From = value
		case "date":
			headers.Date = value
		case "message-id":
			headers.MessageID = value
		case "lines":
			lines, err := strconv.Atoi(value)
			if err == nil {
				headers.Lines = lines
			}
		case "size":
			size, err := strconv.Atoi(value)
			if err == nil {
				headers.Size = size
			}
		}
	}

	return headers, nil
}

// Stat gets the message ID for a specific article number.
// Returns the message ID (without angle brackets).
func (c *NNTPClient) Stat(ctx context.Context, number int) (string, error) {
	if !c.IsConnected() {
		return "", fmt.Errorf("not connected to NNTP server")
	}

	status, responseText, _, err := c.SendCommand(ctx, "STAT", strconv.Itoa(number))
	if err != nil {
		return "", fmt.Errorf("STAT command failed: %w", err)
	}

	if status < 200 || status >= 300 {
		return "", fmt.Errorf("STAT command returned error status %d: %s", status, responseText)
	}

	// STAT response format: "NNN <message-id>"
	responseText = strings.TrimSpace(responseText)
	if strings.HasPrefix(responseText, "<") && strings.HasSuffix(responseText, ">") {
		return responseText[1 : len(responseText)-1], nil
	}

	return responseText, nil
}
