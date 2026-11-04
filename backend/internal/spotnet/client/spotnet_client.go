package client

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/superindexer/indexer/internal/database/models"
)

// SourceType represents the type of a Spotnet source.
type SourceType string

const (
	SourceTypeAPI  SourceType = "api"
	SourceTypeRSS  SourceType = "rss"
	SourceTypeWeb  SourceType = "web"
)

// PaginationConfig holds pagination parameters for API sources.
type PaginationConfig struct {
	PageSize    int    `json:"page_size"`
	CursorField string `json:"cursor_field"`
	PageField   string `json:"page_field"`
	MaxPages    int    `json:"max_pages"`
}

// SpotnetSourceConfig represents the parsed configuration for a Spotnet source.
type SpotnetSourceConfig struct {
	Name          string            `json:"name"`
	SourceType    string            `json:"source_type"`
	URL           string            `json:"url"`
	Method        string            `json:"method"`
	Headers       map[string]string `json:"headers"`
	AuthToken     string            `json:"auth_token"`
	Pagination    PaginationConfig  `json:"pagination"`
	JSONPath      string            `json:"json_path"`
	RSSFeedType   string            `json:"rss_feed_type"`
}

// SpotnetPost represents a post retrieved from a Spotnet source.
type SpotnetPost struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Author      string      `json:"author"`
	Category    string      `json:"category"`
	PostedAt    time.Time   `json:"posted_at"`
	URL         string      `json:"url"`
	RawPost     interface{} `json:"-"`
}

// SpotnetClient handles communication with external Spotnet sources.
type SpotnetClient struct {
	httpClient *http.Client
	maxRetries int
	baseDelay  time.Duration
}

// NewSpotnetClient creates a new SpotnetClient with sensible defaults.
func NewSpotnetClient() *SpotnetClient {
	return &SpotnetClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: 3,
		baseDelay:  1 * time.Second,
	}
}

// ParseSourceConfig decodes the JSON configuration stored on a SpotnetSource
// into a SpotnetSourceConfig.
func ParseSourceConfig(source *models.SpotnetSource) (*SpotnetSourceConfig, error) {
	var cfg SpotnetSourceConfig
	if source.Configuration != "" {
		if err := json.Unmarshal([]byte(source.Configuration), &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse source configuration: %w", err)
		}
	}
	if cfg.SourceType == "" {
		cfg.SourceType = source.SourceType
	}
	if cfg.Name == "" {
		cfg.Name = source.Name
	}
	return &cfg, nil
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// GetPosts retrieves posts from a Spotnet source, optionally filtering by
// an incremental cursor (sinceID or sinceDate).
func (c *SpotnetClient) GetPosts(ctx context.Context, source *models.SpotnetSource, sinceID string, sinceDate time.Time) ([]*SpotnetPost, error) {
	cfg, err := ParseSourceConfig(source)
	if err != nil {
		return nil, err
	}

	if err := c.ValidateSourceConfig(cfg); err != nil {
		return nil, err
	}

	switch SourceType(cfg.SourceType) {
	case SourceTypeAPI:
		return c.fetchAPIPosts(ctx, source, cfg, sinceID, sinceDate)
	case SourceTypeRSS:
		return c.fetchRSSPosts(ctx, source, cfg, sinceID, sinceDate)
	case SourceTypeWeb:
		return c.fetchWebPosts(ctx, source, cfg, sinceID, sinceDate)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", cfg.SourceType)
	}
}

// GetPostsPaginated retrieves a single page of posts from an API source.
func (c *SpotnetClient) GetPostsPaginated(ctx context.Context, source *models.SpotnetSource, page int) ([]*SpotnetPost, error) {
	cfg, err := ParseSourceConfig(source)
	if err != nil {
		return nil, err
	}

	if SourceType(cfg.SourceType) != SourceTypeAPI {
		return nil, fmt.Errorf("pagination only supported for api source type")
	}

	return c.fetchAPIPage(ctx, source, cfg, page)
}

// TestConnection verifies that the client can reach a Spotnet source.
func (c *SpotnetClient) TestConnection(ctx context.Context, source *models.SpotnetSource) error {
	cfg, err := ParseSourceConfig(source)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.applyAuthHeaders(cfg, req)

	resp, err := c.doWithRetry(ctx, func(ctx context.Context) (*http.Response, error) {
		return c.httpClient.Do(req)
	})
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Consume the body so the connection can be reused / closed cleanly.
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))

	return nil
}

// GetAvailableSources returns the list of supported source types.
func (c *SpotnetClient) GetAvailableSources() []string {
	return []string{
		string(SourceTypeAPI),
		string(SourceTypeRSS),
		string(SourceTypeWeb),
	}
}

// ValidateSourceConfig checks that a configuration is well-formed.
func (c *SpotnetClient) ValidateSourceConfig(cfg *SpotnetSourceConfig) error {
	if cfg.Name == "" {
		return fmt.Errorf("name is required")
	}
	if cfg.SourceType == "" {
		return fmt.Errorf("source type is required")
	}
	if cfg.URL == "" {
		return fmt.Errorf("url is required")
	}

	validTypes := c.GetAvailableSources()
	for _, vt := range validTypes {
		if vt == cfg.SourceType {
			return nil
		}
	}
	return fmt.Errorf("invalid source type: %s", cfg.SourceType)
}

// ---------------------------------------------------------------------------
// API source type
// ---------------------------------------------------------------------------

func (c *SpotnetClient) fetchAPIPosts(ctx context.Context, source *models.SpotnetSource, cfg *SpotnetSourceConfig, sinceID string, sinceDate time.Time) ([]*SpotnetPost, error) {
	allPosts, err := c.fetchAllAPIPages(ctx, source, cfg)
	if err != nil {
		return nil, err
	}

	posts := c.filterPosts(allPosts, sinceID, sinceDate)
	c.sortPostsDesc(posts)
	return posts, nil
}

func (c *SpotnetClient) fetchAllAPIPages(ctx context.Context, source *models.SpotnetSource, cfg *SpotnetSourceConfig) ([]*SpotnetPost, error) {
	var allPosts []*SpotnetPost

	maxPages := cfg.Pagination.MaxPages
	if maxPages <= 0 {
		maxPages = 1 // default: no pagination
	}

	for page := 1; page <= maxPages; page++ {
		posts, err := c.fetchAPIPage(ctx, source, cfg, page)
		if err != nil {
			return allPosts, fmt.Errorf("failed to fetch page %d: %w", page, err)
		}
		if len(posts) == 0 {
			break
		}
		allPosts = append(allPosts, posts...)
	}

	return allPosts, nil
}

func (c *SpotnetClient) fetchAPIPage(ctx context.Context, source *models.SpotnetSource, cfg *SpotnetSourceConfig, page int) ([]*SpotnetPost, error) {
	reqURL, err := c.buildAPIURL(cfg.URL, cfg.Pagination, page)
	if err != nil {
		return nil, err
	}

	method := cfg.Method
	if method == "" {
		method = http.MethodGet
	}

	var bodyReader io.Reader
	if method == http.MethodPost {
		jsonBody := c.buildAPIBody(cfg, page)
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.applyAuthHeaders(cfg, req)
	c.applyHeaders(cfg, req)

	resp, err := c.doWithRetry(ctx, func(ctx context.Context) (*http.Response, error) {
		return c.httpClient.Do(req)
	})
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return c.parseAPIResponse(respBody, cfg)
}

func (c *SpotnetClient) buildAPIURL(baseURL string, pg PaginationConfig, page int) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	q := u.Query()
	if pg.PageField != "" {
		q.Set(pg.PageField, fmt.Sprintf("%d", page))
	}
	if pg.CursorField != "" {
		// For cursor-based pagination, we could also support a cursor token
		// in the config; for now we just use page numbers.
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (c *SpotnetClient) buildAPIBody(cfg *SpotnetSourceConfig, page int) []byte {
	params := map[string]interface{}{}
	if cfg.Pagination.PageField != "" {
		params[cfg.Pagination.PageField] = page
	}
	if cfg.Pagination.PageSize > 0 {
		params["page_size"] = cfg.Pagination.PageSize
	}
	data, _ := json.Marshal(params)
	return data
}

func (c *SpotnetClient) parseAPIResponse(body []byte, cfg *SpotnetSourceConfig) ([]*SpotnetPost, error) {
	// Try to unmarshal as JSON array first
	var items []interface{}
	if err := json.Unmarshal(body, &items); err == nil {
		return c.extractPostsFromItems(items, cfg)
	}

	// Try to unmarshal as a JSON object and extract via JSONPath
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil, fmt.Errorf("response is neither a JSON array nor object: %w", err)
	}

	// Handle nested JSON paths like "data.items"
	if cfg.JSONPath != "" {
		val := extractJSONPath(obj, cfg.JSONPath)
		if val != nil {
			switch v := val.(type) {
			case []interface{}:
				return c.extractPostsFromItems(v, cfg)
			case []map[string]interface{}:
				return c.extractPostsFromMaps(v, cfg)
			default:
				// Wrap single item
				return c.extractPostsFromItems([]interface{}{val}, cfg)
			}
		}
	}

	// Fallback: treat the entire object as a single post
	return c.extractPostsFromItems([]interface{}{obj}, cfg)
}

func (c *SpotnetClient) extractPostsFromItems(items []interface{}, cfg *SpotnetSourceConfig) ([]*SpotnetPost, error) {
	posts := make([]*SpotnetPost, 0, len(items))

	for _, item := range items {
		post, err := c.extractPost(item, cfg)
		if err != nil {
			// Log but continue
			continue
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (c *SpotnetClient) extractPost(raw interface{}, cfg *SpotnetSourceConfig) (*SpotnetPost, error) {
	var data map[string]interface{}

	switch v := raw.(type) {
	case map[string]interface{}:
		data = v
	default:
		// Try JSON marshal/unmarshal
		buf, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal raw item: %w", err)
		}
		if err := json.Unmarshal(buf, &data); err != nil {
			return nil, fmt.Errorf("cannot unmarshal raw item: %w", err)
		}
	}

	id := getStringField(data, "id", "ID", "Id", "ID_")
	title := getStringField(data, "title", "Title", "name", "subject", "headline")
	description := getStringField(data, "description", "Description", "body", "desc", "content")
	author := getStringField(data, "author", "Author", "user", "username", "creator")
	category := getStringField(data, "category", "Category", "tag", "tags", "type")
	urlStr := getStringField(data, "url", "URL", "Url", "link", "href", "permalink")

	postedAt, _ := parseTimeField(data)

	return &SpotnetPost{
		ID:          id,
		Title:       title,
		Description: description,
		Author:      author,
		Category:    category,
		PostedAt:    postedAt,
		URL:         urlStr,
		RawPost:     raw,
	}, nil
}

func (c *SpotnetClient) extractPostsFromMaps(items []map[string]interface{}, cfg *SpotnetSourceConfig) ([]*SpotnetPost, error) {
	posts := make([]*SpotnetPost, 0, len(items))
	for _, item := range items {
		post, err := c.extractPost(item, cfg)
		if err != nil {
			continue
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// extractJSONPath walks a dot-separated path in a nested map[string]interface{}
// and returns the value at that path (or nil).
func extractJSONPath(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		default:
			return nil
		}
		if current == nil {
			return nil
		}
	}
	return current
}

// getStringField tries several key names and returns the first non-empty string.
func getStringField(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			if s, ok := val.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// RSS source type
// ---------------------------------------------------------------------------

// rssFeed represents a parsed RSS 2.0 feed.
type rssFeed struct {
	XMLName   xml.Name      `xml:"rss"`
	Version   string        `xml:"version,attr"`
	Title     string        `xml:"channel>title"`
	Link      string        `xml:"channel>link"`
	LinkAtom  string        `xml:"link[rel=self]>href"`
	Description string      `xml:"channel>description"`
	Items     []rssItem     `xml:"channel>item"`
}

// rssItem represents a single item in an RSS feed.
type rssItem struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Author      string    `xml:"author"`
	PubDate     string    `xml:"pubDate"`
	GUID        string    `xml:"guid"`
	Category    string    `xml:"category"`
	Content     string    `xml:"encoded"`
}

// atomFeed represents a parsed Atom feed.
type atomFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string      `xml:"title"`
	Entries []atomEntry `xml:"entry"`
}

// atomEntry represents a single entry in an Atom feed.
type atomEntry struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link[href]"`
	Summary     string    `xml:"summary"`
	Content     string    `xml:"content"`
	Published   string    `xml:"published"`
	Updated     string    `xml:"updated"`
	Author      string    `xml:"author>name"`
	ID          string    `xml:"id"`
	Category    string    `xml:"category[term]"`
}

func (c *SpotnetClient) fetchRSSPosts(ctx context.Context, source *models.SpotnetSource, cfg *SpotnetSourceConfig, sinceID string, sinceDate time.Time) ([]*SpotnetPost, error) {
	resp, err := c.doWithRetry(ctx, func(ctx context.Context) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.URL, nil)
		if err != nil {
			return nil, err
		}
		c.applyAuthHeaders(cfg, req)
		c.applyHeaders(cfg, req)
		return c.httpClient.Do(req)
	})
	if err != nil {
		return nil, fmt.Errorf("RSS fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read RSS body: %w", err)
	}

	var items []rssItem
	feedType := cfg.RSSFeedType
	if feedType == "" {
		feedType = detectFeedType(body)
	}

	switch feedType {
	case "atom":
		atom := &atomFeed{}
		if err := xml.Unmarshal(body, atom); err != nil {
			return nil, fmt.Errorf("failed to parse Atom feed: %w", err)
		}
		items = atomToRSSItems(atom.Entries)
	default:
		rss := &rssFeed{}
		if err := xml.Unmarshal(body, rss); err != nil {
			return nil, fmt.Errorf("failed to parse RSS feed: %w", err)
		}
		items = rss.Items
	}

	posts := make([]*SpotnetPost, 0, len(items))
	for _, item := range items {
		postedAt, _ := time.Parse(time.RFC1123Z, item.PubDate)
		if postedAt.IsZero() {
			postedAt, _ = time.Parse(time.RFC3339, item.PubDate)
		}

		post := &SpotnetPost{
			ID:          item.GUID,
			Title:       item.Title,
			Description: item.Description,
			Author:      item.Author,
			Category:    item.Category,
			PostedAt:    postedAt,
			URL:         item.Link,
			RawPost:     item,
		}
		if post.ID == "" {
			post.ID = post.URL
		}
		posts = append(posts, post)
	}

	posts = c.filterPosts(posts, sinceID, sinceDate)
	c.sortPostsDesc(posts)
	return posts, nil
}

func atomToRSSItems(entries []atomEntry) []rssItem {
	items := make([]rssItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, rssItem{
			Title:       e.Title,
			Link:        e.Link,
			Description: e.Summary,
			Author:      e.Author,
			PubDate:     e.Published,
			GUID:        e.ID,
			Category:    e.Category,
		})
	}
	return items
}

func detectFeedType(body []byte) string {
	// Look for Atom namespace in root element
	if bytes.Contains(body, []byte("xmlns=\"http://www.w3.org/2005/Atom\"")) {
		return "atom"
	}
	if bytes.Contains(body, []byte("<feed")) {
		return "atom"
	}
	return "rss"
}

// ---------------------------------------------------------------------------
// Web source type
// ---------------------------------------------------------------------------

func (c *SpotnetClient) fetchWebPosts(ctx context.Context, source *models.SpotnetSource, cfg *SpotnetSourceConfig, sinceID string, sinceDate time.Time) ([]*SpotnetPost, error) {
	resp, err := c.doWithRetry(ctx, func(ctx context.Context) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.URL, nil)
		if err != nil {
			return nil, err
		}
		c.applyAuthHeaders(cfg, req)
		c.applyHeaders(cfg, req)
		return c.httpClient.Do(req)
	})
	if err != nil {
		return nil, fmt.Errorf("web fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read web page: %w", err)
	}

	// Try JSON first (many APIs return HTML-wrapping JSON or JSON-LD)
	posts, err := c.parseJSONLD(body)
	if err == nil && len(posts) > 0 {
		posts = c.filterPosts(posts, sinceID, sinceDate)
		c.sortPostsDesc(posts)
		return posts, nil
	}

	// Fallback: basic HTML scraping
	posts = c.scrapeHTMLPosts(body, cfg)
	posts = c.filterPosts(posts, sinceID, sinceDate)
	c.sortPostsDesc(posts)
	return posts, nil
}

// parseJSONLD attempts to find and parse JSON-LD structured data in HTML.
func (c *SpotnetClient) parseJSONLD(body []byte) ([]*SpotnetPost, error) {
	// Simple heuristic: find <script type="application/ld+json"> blocks
	idx := bytes.Index(body, []byte(`type="application/ld+json"`))
	if idx == -1 {
		return nil, fmt.Errorf("no JSON-LD found")
	}

	// Find the opening {
	open := bytes.Index(body[idx:], []byte("{"))
	if open == -1 {
		return nil, fmt.Errorf("no JSON-LD content found")
	}

	// Find matching closing } (naive approach)
	depth := 0
	close := open
	for i := open; i < len(body); i++ {
		switch body[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				close = i
				break
			}
		}
	}

	jsonBytes := body[open : close+1]
	var data interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON-LD: %w", err)
	}

	// Try to extract posts from the JSON-LD data
	if arr, ok := data.([]interface{}); ok {
		return c.extractPostsFromItems(arr, &SpotnetSourceConfig{})
	}
	if obj, ok := data.(map[string]interface{}); ok {
		return c.extractPostsFromItems([]interface{}{obj}, &SpotnetSourceConfig{})
	}

	return nil, fmt.Errorf("unsupported JSON-LD structure")
}

// scrapeHTMLPosts performs basic HTML scraping to extract post-like content.
func (c *SpotnetClient) scrapeHTMLPosts(body []byte, cfg *SpotnetSourceConfig) []*SpotnetPost {
	var posts []*SpotnetPost

	// Look for common patterns: <article>, <div class="post">, <li class="item">
	// This is a simplified approach; a full scraper would use goquery or similar.
	content := string(body)

	// Try to find title tags
	for strings.Contains(content, "<title>") {
		start := strings.Index(content, "<title>") + 7
		end := strings.Index(content[start:], "</title>")
		if end == -1 {
			break
		}
		title := strings.TrimSpace(content[start : start+end])
		posts = append(posts, &SpotnetPost{
			Title:  title,
			URL:    cfg.URL,
			PostedAt: time.Now(),
		})
		break // Only get the page title
	}

	return posts
}

// ---------------------------------------------------------------------------
// Filtering & sorting
// ---------------------------------------------------------------------------

func (c *SpotnetClient) filterPosts(posts []*SpotnetPost, sinceID string, sinceDate time.Time) []*SpotnetPost {
	if sinceID == "" && sinceDate.IsZero() {
		return posts
	}

	filtered := make([]*SpotnetPost, 0, len(posts))
	for _, p := range posts {
		if sinceID != "" && p.ID == sinceID {
			// Include this one and stop (assuming newest-first)
			filtered = append(filtered, p)
			break
		}
		if !sinceDate.IsZero() && !p.PostedAt.IsZero() && p.PostedAt.Before(sinceDate) {
			break
		}
		if sinceID != "" && p.ID == sinceID {
			filtered = append(filtered, p)
			break
		}
		filtered = append(filtered, p)
	}

	return filtered
}

func (c *SpotnetClient) sortPostsDesc(posts []*SpotnetPost) {
	for i := 0; i < len(posts); i++ {
		for j := i + 1; j < len(posts); j++ {
			if posts[j].PostedAt.After(posts[i].PostedAt) {
				posts[i], posts[j] = posts[j], posts[i]
			}
		}
	}
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

func (c *SpotnetClient) applyAuthHeaders(cfg *SpotnetSourceConfig, req *http.Request) {
	if cfg.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)
	}
}

func (c *SpotnetClient) applyHeaders(cfg *SpotnetSourceConfig, req *http.Request) {
	if cfg.Headers == nil {
		return
	}
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}
}

// doWithRetry executes a function with exponential backoff retry.
func (c *SpotnetClient) doWithRetry(ctx context.Context, fn func(ctx context.Context) (*http.Response, error)) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			delay := c.baseDelay * time.Duration(1<<uint(attempt-1)) // 1s, 2s, 4s...
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		resp, err := fn(ctx)
		if err != nil {
			lastErr = err
			continue
		}

		// Retry on server errors
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", c.maxRetries, lastErr)
}

// ---------------------------------------------------------------------------
// Time parsing helpers
// ---------------------------------------------------------------------------

func parseTimeField(data map[string]interface{}) (time.Time, error) {
	timeFields := []string{
		"posted_at", "created_at", "updated_at", "published_at",
		"date", "timestamp", "pub_date", "pubDate", "published",
	}

	for _, field := range timeFields {
		if val, ok := data[field]; ok {
			switch v := val.(type) {
			case string:
				return parseTime(v)
			case float64:
				return time.Unix(int64(v), 0).UTC(), nil
			case json.Number:
				if ts, err := v.Int64(); err == nil {
					return time.Unix(ts, 0).UTC(), nil
				}
			}
		}
	}
	return time.Time{}, fmt.Errorf("no time field found")
}

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}

	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}
