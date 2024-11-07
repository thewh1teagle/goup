package updater

import "io"

type ProgressCallback func(current int64, total int64)

type ProgressWriter struct {
	Writer           io.Writer
	ProgressCallback ProgressCallback
	TotalSize        int64
	CurrentSize      int64
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.Writer.Write(p)
	if err == nil {
		pw.CurrentSize += int64(n)
		if pw.ProgressCallback != nil {
			pw.ProgressCallback(pw.CurrentSize, pw.TotalSize)
		}
	}
	return n, err
}
