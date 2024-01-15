package grpc

import (
	"fmt"
	"io"
)

func ForwardStreamLogs(w io.Writer, stream JobService_StreamLogsClient) error {
	for {
		resp, err := stream.Recv() // it's blocking operation, but it will be released, when stream will be closed/canceled
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		fmt.Fprintf(w, "%s", resp.Output) // assumption that it UTF-8
	}
}
