// Package sources hosts pluggable URL resolvers that turn a "page" URL into one
// or more directly-downloadable media URLs before handing them to the core
// download engine.
package sources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

// ErrNotImplemented is returned by scaffolded resolvers.
var ErrNotImplemented = errors.New("sources: not implemented yet")

// ErrYtDlpMissing is returned when the yt-dlp binary cannot be found on PATH.
var ErrYtDlpMissing = errors.New("yt-dlp not found on PATH; install it from https://github.com/yt-dlp/yt-dlp")

// Media is a single resolved, directly-downloadable target.
type Media struct {
	URL      string
	Filename string
	Headers  map[string]string
}

// Resolver turns an input URL into concrete downloadable media.
type Resolver interface {
	Supports(rawURL string) bool
	Resolve(ctx context.Context, rawURL string) ([]Media, error)
}

// YtDlp resolves video/stream pages using the external yt-dlp binary.
type YtDlp struct {
	// BinaryPath is the path to the yt-dlp executable. Empty means "yt-dlp" on PATH.
	BinaryPath string
}

func (y YtDlp) bin() string {
	if y.BinaryPath != "" {
		return y.BinaryPath
	}
	return "yt-dlp"
}

// directExt lists file extensions that are already direct downloads, so we can
// skip yt-dlp for plain files.
var directExt = map[string]bool{
	".zip": true, ".tar": true, ".gz": true, ".7z": true, ".rar": true,
	".iso": true, ".exe": true, ".dmg": true, ".pkg": true, ".deb": true,
	".pdf": true, ".apk": true, ".bin": true, ".msi": true,
}

// Supports reports whether yt-dlp is a reasonable resolver for the URL. It
// returns true for http(s) URLs that are not obviously a direct file download.
func (y YtDlp) Supports(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}
	lower := strings.ToLower(u.Path)
	for ext := range directExt {
		if strings.HasSuffix(lower, ext) {
			return false
		}
	}
	return true
}

// ytInfo is the subset of yt-dlp's JSON (-J) output that we consume.
type ytInfo struct {
	URL                string `json:"url"`
	Ext                string `json:"ext"`
	Title              string `json:"title"`
	RequestedDownloads []struct {
		URL         string            `json:"url"`
		Ext         string            `json:"ext"`
		HTTPHeaders map[string]string `json:"http_headers"`
	} `json:"requested_downloads"`
	HTTPHeaders map[string]string `json:"http_headers"`
}

// Resolve runs `yt-dlp -J --no-playlist <url>` and extracts a direct media URL
// (and any required HTTP headers, e.g. Referer/User-Agent).
func (y YtDlp) Resolve(ctx context.Context, rawURL string) ([]Media, error) {
	if _, err := exec.LookPath(y.bin()); err != nil {
		return nil, ErrYtDlpMissing
	}

	cmd := exec.CommandContext(ctx, y.bin(), "-J", "--no-playlist", "--no-warnings", rawURL)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %w", err)
	}

	var info ytInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("could not parse yt-dlp output: %w", err)
	}

	title := strings.TrimSpace(info.Title)

	// Prefer requested_downloads (the format yt-dlp would actually fetch).
	if len(info.RequestedDownloads) > 0 {
		var media []Media
		for _, rd := range info.RequestedDownloads {
			if rd.URL == "" {
				continue
			}
			media = append(media, Media{
				URL:      rd.URL,
				Filename: filename(title, rd.Ext),
				Headers:  rd.HTTPHeaders,
			})
		}
		if len(media) > 0 {
			return media, nil
		}
	}

	if info.URL != "" {
		m := Media{URL: info.URL, Filename: filename(title, info.Ext), Headers: info.HTTPHeaders}
		return []Media{m}, nil
	}

	return nil, errors.New("yt-dlp returned no downloadable URL")
}

func filename(title, ext string) string {
	if title == "" {
		return ""
	}
	safe := strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			return '_'
		}
		return r
	}, title)
	if ext != "" {
		return safe + "." + ext
	}
	return safe
}

var _ Resolver = YtDlp{}
