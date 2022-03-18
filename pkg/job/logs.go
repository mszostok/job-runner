package job

import (
	"context"
	"fmt"
	"io"
)

func ForwardStreamLogs(ctx context.Context, w io.Writer, stream *StreamLogsOutput) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-stream.Output:
			if !ok { // out closed, no need to watch for new messages
				return nil
			}
			fmt.Fprintf(w, "%s", msg)
		case err := <-stream.Error:
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return fmt.Errorf("while streaming logs %w", err)
			}
		}
	}
}
