package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/javaman/go-metrics/internal/tools"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
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
		decompressingReader, err := newCompressReader(b)
		if err == nil {
			c.Request().Body = decompressingReader
			return next(c)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
}

type bufferedWriter struct {
	w      http.ResponseWriter
	buffer []byte
	status *int
}

func newBufferedWriter(w http.ResponseWriter) *bufferedWriter {
	return &bufferedWriter{
		w:      w,
		buffer: nil,
		status: nil,
	}
}

func (b *bufferedWriter) Header() http.Header {
	return b.w.Header()
}

func (b *bufferedWriter) Write(p []byte) (int, error) {
	b.buffer = append(b.buffer, p...)
	return len(p), nil
}

func (b *bufferedWriter) WriteHeader(statusCode int) {
	//fmt.Println(b.buffer)
	//fmt.Println("Header is written")
	//b.w.WriteHeader(statusCode)
	b.status = &statusCode
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

func CompressDecompress(next echo.HandlerFunc) echo.HandlerFunc {
	return Decompress(Compress(next))
}

func VerifyHash(key string) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			hash := c.Request().Header.Get("HashSHA256")
			if len(hash) > 0 {
				body, err := io.ReadAll(c.Request().Body)
				c.Request().Body = io.NopCloser(bytes.NewBuffer(body))
				if err != nil {
					return c.NoContent(http.StatusInternalServerError)
				} else if !tools.AreEquals(hash, tools.ComputeSign(body, key)) {
					return c.NoContent(http.StatusBadRequest)
				}
			}
			return next(c)
		}
	}
}

func AppendHash(key string) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rw := c.Response().Writer
			cw := newBufferedWriter(rw)
			c.Response().Writer = cw
			defer func() {
				rw.Header().Set("HashSHA256", tools.ComputeSign(cw.buffer, key))
				rw.WriteHeader(*cw.status)
				rw.Write(cw.buffer)
				c.Response().Writer = rw
			}()
			return next(c)
		}
	}
}

func Logger() echo.MiddlewareFunc {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:          true,
		LogMethod:       true,
		LogLatency:      true,
		LogStatus:       true,
		LogResponseSize: true,
		LogValuesFunc: func(e echo.Context, v middleware.RequestLoggerValues) error {
			sugar.Infow("request",
				zap.String("URI", v.URI),
				zap.String("method", v.Method),
				zap.Int64("latency", v.Latency.Nanoseconds()),
			)
			sugar.Infow("response",
				zap.Int("status", v.Status),
				zap.Int64("size", v.ResponseSize),
			)
			return nil
		},
	})
}
