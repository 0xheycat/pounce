package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/0xheycat/pounce/internal/model"
	"github.com/0xheycat/pounce/internal/ratelimit"
)

// readChunk is the buffer size used when streaming a segment to disk.
const readChunk = 32 * 1024

// downloadSegment streams one byte-range into the target file using HTTP Range
// requests. It writes at absolute offsets with WriteAt, so all segments can
// share a single *os.File concurrently. Progress is reported via onBytes.
func (m *Manager) downloadSegment(
	ctx context.Context,
	d *model.Download,
	seg *model.Segment,
	f *os.File,
	lim *ratelimit.Limiter,
	onBytes func(int),
) error {
	start := seg.Start + seg.Downloaded
	if seg.End >= 0 && start > seg.End {
		return nil // already complete
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.URL, nil)
	if err != nil {
		return err
	}
	if d.Resumable {
		if seg.End >= 0 {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, seg.End))
		} else {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", start))
		}
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("segment %d: unexpected status %s", seg.Index, resp.Status)
	}

	buf := make([]byte, readChunk)
	offset := start
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		nr, rerr := resp.Body.Read(buf)
		if nr > 0 {
			if werr := lim.WaitN(ctx, nr); werr != nil {
				return werr
			}
			nw, werr := f.WriteAt(buf[:nr], offset)
			if werr != nil {
				return werr
			}
			offset += int64(nw)
			seg.Downloaded += int64(nw)
			onBytes(nw)
		}
		if rerr == io.EOF {
			return nil
		}
		if rerr != nil {
			return rerr
		}
	}
}
