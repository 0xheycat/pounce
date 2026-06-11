// Package download contains the Pounce download engine: probing, segmented
// multi-connection transfers, persistent resume, throttling and queueing.
package download

import (
	"context"
	"fmt"
	"math/rand"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0xheycat/pounce/internal/model"
	"github.com/0xheycat/pounce/internal/ratelimit"
	"github.com/0xheycat/pounce/internal/store"
)

// Manager owns every download and coordinates the worker goroutines.
type Manager struct {
	mu           sync.Mutex
	entries      map[string]*entry
	store        *store.Store
	client       *http.Client
	emit         func(*model.Download)
	defaultConns int
}

type entry struct {
	d       *model.Download
	cancel  context.CancelFunc
	limiter *ratelimit.Limiter
	running bool
}

// NewManager builds a Manager. emit is called whenever a download changes so
// the API layer can push live updates (it may be nil).
func NewManager(st *store.Store, emit func(*model.Download)) *Manager {
	return &Manager{
		entries:      make(map[string]*entry),
		store:        st,
		client:       &http.Client{},
		emit:         emit,
		defaultConns: 8,
	}
}

// snapshot returns a deep copy of d with every atomically-updated counter read
// safely. Callers must hold m.mu, which guards the status/speed/error fields.
// Working from the copy lets us marshal and broadcast a download without racing
// the segment worker goroutines.
func (m *Manager) snapshot(d *model.Download) *model.Download {
	cp := &model.Download{
		ID:          d.ID,
		URL:         d.URL,
		Filename:    d.Filename,
		Dir:         d.Dir,
		TotalSize:   d.TotalSize,
		Downloaded:  atomic.LoadInt64(&d.Downloaded),
		Status:      d.Status,
		Connections: d.Connections,
		Resumable:   d.Resumable,
		SpeedLimit:  d.SpeedLimit,
		Error:       d.Error,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
		Speed:       d.Speed,
	}
	cp.Segments = make([]model.Segment, len(d.Segments))
	for i := range d.Segments {
		cp.Segments[i] = model.Segment{
			Index:      d.Segments[i].Index,
			Start:      d.Segments[i].Start,
			End:        d.Segments[i].End,
			Downloaded: atomic.LoadInt64(&d.Segments[i].Downloaded),
		}
	}
	return cp
}

func (m *Manager) notify(d *model.Download) {
	m.mu.Lock()
	cp := m.snapshot(d)
	m.mu.Unlock()
	if m.emit != nil {
		m.emit(cp)
	}
	_ = m.store.Save(cp)
}

// LoadExisting restores downloads from disk. Anything that was running when the
// process stopped is marked paused so the user can resume it.
func (m *Manager) LoadExisting() error {
	ds, err := m.store.LoadAll()
	if err != nil {
		return err
	}
	m.mu.Lock()
	for _, d := range ds {
		if d.Status == model.StatusRunning {
			d.Status = model.StatusPaused
		}
		d.Speed = 0
		m.entries[d.ID] = &entry{d: d, limiter: ratelimit.New(d.SpeedLimit)}
	}
	m.mu.Unlock()
	m.cleanOrphans()
	return nil
}

// cleanOrphans deletes stray ".pdownload" part files that no longer belong to a
// known download - for example data left behind by a crash, or by a cancel
// that raced the worker goroutine. Only directories that downloads actually use
// are scanned, so unrelated files are never touched.
func (m *Manager) cleanOrphans() {
	m.mu.Lock()
	valid := make(map[string]struct{}, len(m.entries))
	dirs := map[string]struct{}{DefaultDownloadDir(): {}}
	for _, e := range m.entries {
		valid[filepath.Join(e.d.Dir, e.d.Filename+".pdownload")] = struct{}{}
		dirs[e.d.Dir] = struct{}{}
	}
	m.mu.Unlock()

	for dir := range dirs {
		ents, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, ent := range ents {
			if ent.IsDir() || filepath.Ext(ent.Name()) != ".pdownload" {
				continue
			}
			full := filepath.Join(dir, ent.Name())
			if _, ok := valid[full]; !ok {
				_ = os.Remove(full)
			}
		}
	}
}

// Add probes the URL, creates a download and returns it (queued, not started).
func (m *Manager) Add(rawurl, dir string, conns int, speedLimit int64) (*model.Download, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	size, resumable, filename, err := m.probe(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	if conns <= 0 {
		conns = m.defaultConns
	}
	if !resumable || size <= 0 {
		conns = 1
	}
	if dir == "" {
		dir = DefaultDownloadDir()
	}

	d := &model.Download{
		ID:          newID(),
		URL:         rawurl,
		Filename:    uniqueName(dir, filename),
		Dir:         dir,
		TotalSize:   size,
		Status:      model.StatusQueued,
		Connections: conns,
		Resumable:   resumable,
		SpeedLimit:  speedLimit,
		Segments:    buildSegments(size, conns),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	m.mu.Lock()
	m.entries[d.ID] = &entry{d: d, limiter: ratelimit.New(speedLimit)}
	cp := m.snapshot(d)
	m.mu.Unlock()
	m.notify(d)
	return cp, nil
}

// Start begins (or resumes) a download.
func (m *Manager) Start(id string) error {
	m.mu.Lock()
	e, ok := m.entries[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("download %q not found", id)
	}
	if e.running {
		m.mu.Unlock()
		return nil
	}
	if e.d.Status == model.StatusCompleted {
		m.mu.Unlock()
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	e.running = true
	d := e.d
	lim := e.limiter
	m.mu.Unlock()

	go m.run(ctx, e, d, lim)
	return nil
}

// Pause stops a download but keeps partial data so it can resume later.
func (m *Manager) Pause(id string) error {
	m.mu.Lock()
	e, ok := m.entries[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("download %q not found", id)
	}
	e.d.Status = model.StatusPaused
	e.d.Speed = 0
	e.d.UpdatedAt = time.Now()
	cancel := e.cancel
	d := e.d
	m.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	m.notify(d)
	return nil
}

// SetSpeed updates the per-download speed limit live (bytes/sec, 0 = unlimited).
func (m *Manager) SetSpeed(id string, limit int64) error {
	m.mu.Lock()
	e, ok := m.entries[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("download %q not found", id)
	}
	e.d.SpeedLimit = limit
	d := e.d
	lim := e.limiter
	m.mu.Unlock()

	lim.SetRate(limit)
	m.notify(d)
	return nil
}

// Remove cancels a download and deletes its metadata. When deleteFile is true
// the partial file is removed too.
func (m *Manager) Remove(id string, deleteFile bool) error {
	m.mu.Lock()
	e, ok := m.entries[id]
	if ok {
		e.d.Status = model.StatusCanceled
		if e.cancel != nil {
			e.cancel()
		}
		delete(m.entries, id)
	}
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("download %q not found", id)
	}
	_ = m.store.Delete(id)
	if deleteFile {
		_ = os.Remove(filepath.Join(e.d.Dir, e.d.Filename+".pdownload"))
	}
	return nil
}

// List returns a snapshot of all downloads.
func (m *Manager) List() []*model.Download {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*model.Download, 0, len(m.entries))
	for _, e := range m.entries {
		out = append(out, m.snapshot(e.d))
	}
	return out
}

func (m *Manager) run(ctx context.Context, e *entry, d *model.Download, lim *ratelimit.Limiter) {
	m.mu.Lock()
	d.Status = model.StatusRunning
	d.Error = ""
	d.UpdatedAt = time.Now()
	m.mu.Unlock()

	// Non-resumable transfers must always restart from zero.
	if !d.Resumable {
		atomic.StoreInt64(&d.Downloaded, 0)
		for i := range d.Segments {
			atomic.StoreInt64(&d.Segments[i].Downloaded, 0)
		}
	}
	m.notify(d)

	if err := os.MkdirAll(d.Dir, 0o755); err != nil {
		m.fail(e, d, err)
		return
	}
	partPath := filepath.Join(d.Dir, d.Filename+".pdownload")
	f, err := os.OpenFile(partPath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		m.fail(e, d, err)
		return
	}
	if d.TotalSize > 0 {
		_ = f.Truncate(d.TotalSize)
	}

	stop := make(chan struct{})
	go m.progressLoop(d, stop)

	var wg sync.WaitGroup
	var firstErr atomic.Value
	for i := range d.Segments {
		seg := &d.Segments[i]
		if seg.End >= 0 && atomic.LoadInt64(&seg.Downloaded) >= (seg.End-seg.Start+1) {
			continue // segment already finished
		}
		wg.Add(1)
		go func(seg *model.Segment) {
			defer wg.Done()
			if err := m.downloadSegment(ctx, d, seg, f, lim, func(n int) {
				atomic.AddInt64(&d.Downloaded, int64(n))
			}); err != nil && firstErr.Load() == nil {
				firstErr.Store(err)
			}
		}(seg)
	}
	wg.Wait()
	close(stop)
	_ = f.Sync()
	_ = f.Close()

	m.mu.Lock()
	e.running = false
	m.mu.Unlock()

	if ctx.Err() != nil {
		m.mu.Lock()
		canceled := d.Status == model.StatusCanceled
		if !canceled {
			d.Status = model.StatusPaused
		}
		d.Speed = 0
		d.UpdatedAt = time.Now()
		m.mu.Unlock()
		if canceled {
			// The worker owns the part file, so it deletes it last - after the
			// segment goroutines have stopped - to avoid racing a recreate with
			// Remove(), which already removed the metadata. Emit the final state so
			// SSE clients drop the row, but do not re-persist it.
			_ = os.Remove(partPath)
			if m.emit != nil {
				m.mu.Lock()
				cp := m.snapshot(d)
				m.mu.Unlock()
				m.emit(cp)
			}
			return
		}
		m.notify(d)
		return
	}
	if v := firstErr.Load(); v != nil {
		m.fail(e, d, v.(error))
		return
	}

	if rerr := os.Rename(partPath, filepath.Join(d.Dir, d.Filename)); rerr != nil {
		m.fail(e, d, rerr)
		return
	}
	if d.TotalSize > 0 {
		atomic.StoreInt64(&d.Downloaded, d.TotalSize)
	}
	m.mu.Lock()
	d.Status = model.StatusCompleted
	d.Speed = 0
	d.UpdatedAt = time.Now()
	m.mu.Unlock()
	m.notify(d)
}

func (m *Manager) progressLoop(d *model.Download, stop <-chan struct{}) {
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	last := atomic.LoadInt64(&d.Downloaded)
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			cur := atomic.LoadInt64(&d.Downloaded)
			m.mu.Lock()
			d.Speed = (cur - last) * 2 // sampled every 0.5s -> bytes/sec
			d.UpdatedAt = time.Now()
			m.mu.Unlock()
			last = cur
			m.notify(d)
		}
	}
}

func (m *Manager) fail(e *entry, d *model.Download, err error) {
	m.mu.Lock()
	e.running = false
	d.Status = model.StatusError
	d.Error = err.Error()
	d.Speed = 0
	d.UpdatedAt = time.Now()
	m.mu.Unlock()
	m.notify(d)
}

// probe discovers the size, range support and a filename for a URL.
func (m *Manager) probe(ctx context.Context, rawurl string) (size int64, resumable bool, filename string, err error) {
	if req, e := http.NewRequestWithContext(ctx, http.MethodHead, rawurl, nil); e == nil {
		if resp, e2 := m.client.Do(req); e2 == nil {
			resp.Body.Close()
			size = resp.ContentLength
			resumable = strings.EqualFold(resp.Header.Get("Accept-Ranges"), "bytes")
			filename = filenameFrom(resp, rawurl)
			if size > 0 {
				return size, resumable, filename, nil
			}
		}
	}

	// Fallback: a tiny ranged GET tells us both size and range support.
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, rawurl, nil)
	if e != nil {
		return 0, false, "download", e
	}
	req.Header.Set("Range", "bytes=0-0")
	resp, e := m.client.Do(req)
	if e != nil {
		return 0, false, "download", e
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusPartialContent {
		resumable = true
		if cr := resp.Header.Get("Content-Range"); cr != "" {
			if i := strings.LastIndex(cr, "/"); i >= 0 {
				if v, perr := strconv.ParseInt(strings.TrimSpace(cr[i+1:]), 10, 64); perr == nil {
					size = v
				}
			}
		}
	} else {
		size = resp.ContentLength
	}
	return size, resumable, filenameFrom(resp, rawurl), nil
}

func buildSegments(size int64, conns int) []model.Segment {
	if size <= 0 || conns <= 1 {
		end := int64(-1)
		if size > 0 {
			end = size - 1
		}
		seg := model.Segment{Index: 0, Start: 0, End: end}
		return []model.Segment{seg}
	}
	segs := make([]model.Segment, 0, conns)
	chunk := size / int64(conns)
	var start int64
	for i := 0; i < conns; i++ {
		end := start + chunk - 1
		if i == conns-1 {
			end = size - 1
		}
		segs = append(segs, model.Segment{Index: i, Start: start, End: end})
		start = end + 1
	}
	return segs
}

func filenameFrom(resp *http.Response, rawurl string) string {
	if resp != nil {
		if cd := resp.Header.Get("Content-Disposition"); cd != "" {
			if _, params, err := mime.ParseMediaType(cd); err == nil {
				if fn := params["filename"]; fn != "" {
					return fn
				}
			}
		}
	}
	if u, err := url.Parse(rawurl); err == nil {
		if base := path.Base(u.Path); base != "" && base != "." && base != "/" {
			return base
		}
	}
	return "download"
}

func uniqueName(dir, name string) string {
	if _, err := os.Stat(filepath.Join(dir, name)); os.IsNotExist(err) {
		return name
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	for i := 1; ; i++ {
		cand := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if _, err := os.Stat(filepath.Join(dir, cand)); os.IsNotExist(err) {
			return cand
		}
	}
}

// DefaultDownloadDir returns ~/Downloads/Pounce (or the working dir as a
// fallback).
func DefaultDownloadDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, "Downloads", "Pounce")
}

func newID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36) + strconv.Itoa(rand.Intn(1000))
}
