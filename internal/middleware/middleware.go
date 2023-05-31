package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	contentEncodingGzip = "gzip"
)

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func Decompress(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Header.Get("Content-Encoding") != contentEncodingGzip {
			return next(c)
		}
		b := c.Request().Body
		if decompressingReader, err := newCompressReader(b); err == nil {
			c.Request().Body = decompressingReader
			return next(c)
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
}

type compressWriter struct {
	w         http.ResponseWriter
	zw        *gzip.Writer
	wroteBody bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:         w,
		zw:        gzip.NewWriter(w),
		wroteBody: false,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	c.wroteBody = true
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.Header().Del(echo.HeaderContentLength)
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", contentEncodingGzip)
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}
func Compress(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if strings.Contains(c.Request().Header.Get("Accept-Encoding"), contentEncodingGzip) {
			rw := c.Response().Writer
			cw := newCompressWriter(rw)
			cw.zw.Reset(rw)
			defer func() {
				if !cw.wroteBody {
					if c.Response().Header().Get("Content-Encoding") == contentEncodingGzip {
						c.Response().Header().Del("Content-Encoding")
					}
					c.Response().Writer = rw
					cw.zw.Reset(io.Discard)
				}
				cw.zw.Close()
			}()
			c.Response().Writer = cw
		}
		return next(c)
	}
}
