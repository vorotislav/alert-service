package middlewares

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p) //nolint:wrapcheck
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < http.StatusMultipleChoices {
		c.w.Header().Set("Content-Encoding", "gzip")
	}

	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	return c.zw.Close() //nolint:wrapcheck
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *compressReader) Read(p []byte) (int, error) {
	return c.zr.Read(p) //nolint:wrapcheck
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err //nolint:wrapcheck
	}

	return c.zr.Close() //nolint:wrapcheck
}

func CompressMiddleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		writer := w

		acceptEncoding := r.Header.Get("Accept-Encoding")

		if strings.Contains(acceptEncoding, "gzip") {
			w.Header().Set("Accept-Encoding", "gzip")
			cw := newCompressWriter(w)

			writer = cw

			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		if strings.Contains(contentEncoding, "gzip") {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("cannot create decompressor: %s", err.Error()), http.StatusBadRequest)

				return
			}

			r.Body = cr

			defer cr.Close()
		}

		h.ServeHTTP(writer, r)
	}

	return http.HandlerFunc(fn)
}
