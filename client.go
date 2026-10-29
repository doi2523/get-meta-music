package metamusic

import (
	"context"
	"net/http"
	"time"

	"github.com/doi2523/get-meta-music/pkg/music"
)

// Logger defines the minimal logging interface.
type Logger = music.Logger

// Client provides an interface for extracting song metadata with custom configurations.
type Client struct {
	timeout time.Duration
	proxy   string
	client  *http.Client
	logger  music.Logger
}

// Option is a functional option for configuring a Client.
type Option func(*Client)

// WithTimeout sets a custom request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.timeout = d
	}
}

// WithProxy sets an HTTP, HTTPS, or SOCKS5 proxy URL.
func WithProxy(proxyURL string) Option {
	return func(c *Client) {
		c.proxy = proxyURL
	}
}

// WithHTTPClient provides a custom http.Client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

// WithLogger sets a custom logger.
func WithLogger(l music.Logger) Option {
	return func(c *Client) {
		c.logger = l
	}
}

// WithSilent suppresses all internal logging.
func WithSilent() Option {
	return func(c *Client) {
		c.logger = music.NopLogger
	}
}

// NewClient creates a new Client configured with the provided options.
func NewClient(opts ...Option) *Client {
	c := &Client{
		timeout: 15 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}

	if c.client != nil {
		music.SetHTTPClient(c.client)
	}
	if c.timeout > 0 {
		music.SetTimeout(c.timeout)
	}
	if c.proxy != "" {
		_ = music.SetProxyURL(c.proxy)
	}
	if c.logger != nil {
		music.SetLogger(c.logger)
	}

	return c
}

// Get extracts metadata from either a Spotify or Apple Music URL.
func (c *Client) Get(rawURL string) (*SongMetadata, error) {
	return music.GetSongMetadata(rawURL)
}

// GetWithContext extracts metadata respecting context cancellation or deadline.
func (c *Client) GetWithContext(ctx context.Context, rawURL string) (*SongMetadata, error) {
	type result struct {
		meta *SongMetadata
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		m, err := music.GetSongMetadata(rawURL)
		ch <- result{meta: m, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		return res.meta, res.err
	}
}

// GetSpotify extracts song metadata specifically for Spotify URLs.
func (c *Client) GetSpotify(rawURL string) (*SongMetadata, error) {
	return music.GetSpotifyMetadata(rawURL)
}

// GetSpotifyWithContext extracts Spotify metadata with context.
func (c *Client) GetSpotifyWithContext(ctx context.Context, rawURL string) (*SongMetadata, error) {
	type result struct {
		meta *SongMetadata
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		m, err := music.GetSpotifyMetadata(rawURL)
		ch <- result{meta: m, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		return res.meta, res.err
	}
}

// GetAppleMusic extracts song metadata specifically for Apple Music URLs.
func (c *Client) GetAppleMusic(rawURL string) (*SongMetadata, error) {
	return music.GetAppleMusicMetadata(rawURL)
}

// GetAppleMusicWithContext extracts Apple Music metadata with context.
func (c *Client) GetAppleMusicWithContext(ctx context.Context, rawURL string) (*SongMetadata, error) {
	type result struct {
		meta *SongMetadata
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		m, err := music.GetAppleMusicMetadata(rawURL)
		ch <- result{meta: m, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		return res.meta, res.err
	}
}

// GetISRC resolves the ISRC code and cross-platform URLs for a song link.
func (c *Client) GetISRC(rawURL string) (*ISRCData, error) {
	return music.GetISRCFromURL(rawURL)
}

// GetISRCWithContext resolves ISRC with context.
func (c *Client) GetISRCWithContext(ctx context.Context, rawURL string) (*ISRCData, error) {
	type result struct {
		data *ISRCData
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		d, err := music.GetISRCFromURL(rawURL)
		ch <- result{data: d, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		return res.data, res.err
	}
}
