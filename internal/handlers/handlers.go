package handlers

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/javaman/go-metrics/internal/model"
	"github.com/javaman/go-metrics/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

const (
	ContentEncodingGzip = "gzip"
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
		if c.Request().Header.Get("Content-Encoding") != ContentEncodingGzip {
			return next(c)
		} else {
			b := c.Request().Body
			defer b.Close()
			if decompressingReader, err := newCompressReader(b); err == nil {
				defer decompressingReader.Close()
				c.Request().Body = decompressingReader
				return next(c)
			} else {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
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
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

func Compress(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if strings.Contains(c.Request().Header.Get("Accept-Encoding"), ContentEncodingGzip) {
			c.Response().Header().Set("Content-Encoding", ContentEncodingGzip)
			rw := c.Response().Writer
			cw := newCompressWriter(rw)
			cw.zw.Reset(rw)
			defer func() {
				if !cw.wroteBody {
					if c.Response().Header().Get("Content-Encoding") == ContentEncodingGzip {
						c.Response().Header().Del("Content-Encoding")
					}
					c.Response().Writer = rw
					cw.zw.Reset(io.Discard)
				}
				cw.zw.Close()
			}()
			c.Response().Writer = cw
			return next(c)
		} else {
			return next(c)
		}
	}
}

func BadRequest(c echo.Context) error {
	return c.NoContent(http.StatusBadRequest)
}

func NotFound(c echo.Context) error {
	return c.NoContent(http.StatusNotFound)
}

func ValueGauge(s services.MetricsService) func(echo.Context) error {
	return func(c echo.Context) error {
		measureName := c.Param("measureName")
		if value, found := s.GetGauge(measureName); found {
			return c.String(http.StatusOK, strconv.FormatFloat(value, 'f', -1, 64))
		} else {
			return NotFound(c)
		}
	}
}

func UpdateGauge(s services.MetricsService) func(echo.Context) error {
	return func(c echo.Context) error {
		measureName := c.Param("measureName")
		if strings.TrimSpace(measureName) == "" {
			return NotFound(c)
		}
		if measureValue, err := strconv.ParseFloat(c.Param("measureValue"), 64); err == nil {
			s.SaveGauge(measureName, measureValue)
			return c.NoContent(http.StatusOK)
		} else {
			return BadRequest(c)
		}
	}
}

func ValueCounter(s services.MetricsService) func(echo.Context) error {
	return func(c echo.Context) error {
		measureName := c.Param("measureName")
		fmt.Println(c.ParamNames())
		fmt.Println(c.ParamValues())
		if value, found := s.GetCounter(measureName); found {
			return c.String(http.StatusOK, fmt.Sprintf("%d", value))
		} else {
			return NotFound(c)
		}
	}
}
func UpdateCounter(s services.MetricsService) func(echo.Context) error {
	return func(c echo.Context) error {
		metricName := c.Param("measureName")
		if strings.TrimSpace(metricName) == "" {
			return NotFound(c)
		}
		if metricValue, err := strconv.ParseInt(c.Param("measureValue"), 10, 64); err == nil {
			s.SaveCounter(metricName, metricValue)
			return c.NoContent(http.StatusOK)
		} else {
			return BadRequest(c)
		}
	}
}

func ListAll(s services.MetricsService) func(echo.Context) error {
	return func(e echo.Context) error {
		var b strings.Builder
		b.WriteString("<html><head><title>AllMetrics</title></head><body><table>")
		s.AllGauges(func(n string, v float64) {
			b.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%f</td></tr>", n, v))
		})
		s.AllCounters(func(n string, v int64) {
			b.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>", n, v))
		})
		b.WriteString("</table></body></html>")
		return e.HTML(http.StatusOK, b.String())
	}
}

func Update(s services.MetricsService) func(echo.Context) error {
	return func(c echo.Context) error {
		var m model.Metrics
		err := json.NewDecoder(c.Request().Body).Decode(&m)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		if res, err := s.Save(&m); err != nil {
			switch err {
			case services.ErrIDRequired:
				return NotFound(c)
			case services.ErrInvalidMType:
				fallthrough
			case services.ErrDeltaRequired:
				fallthrough
			case services.ErrValueRequired:
				fallthrough
			default:
				return BadRequest(c)
			}
		} else {
			//
			return c.JSON(http.StatusOK, res)
		}
	}
}

func Value(s services.MetricsService) func(echo.Context) error {
	return func(c echo.Context) error {
		var m model.Metrics
		err := json.NewDecoder(c.Request().Body).Decode(&m)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		if res, err := s.Value(&m); err != nil {
			switch err {
			case services.ErrIDNotFound:
				return c.NoContent(http.StatusNotFound)
			default:
				return BadRequest(c)
			}
		} else {
			return c.JSON(http.StatusOK, res)
		}
	}
}

func New(service services.MetricsService) *echo.Echo {
	e := echo.New()

	e.GET("/", ListAll(service))

	e.GET("/value/counter/:measureName", ValueCounter(service))
	e.GET("/value/gauge/:measureName", ValueGauge(service))
	e.POST("/value/", Value(service))

	e.GET("/update/*", func(c echo.Context) error { return c.NoContent(http.StatusMethodNotAllowed) })
	e.POST("/update/:measureType/*", BadRequest)

	e.POST("/update/counter/:measureName/:measureValue", UpdateCounter(service))
	e.POST("/update/counter/", NotFound)
	e.POST("/update/gauge/:measureName/:measureValue", UpdateGauge(service))
	e.POST("/update/gauge/", NotFound)
	e.POST("/update/", Update(service))

	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
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
	}))
	e.Use(Compress)
	e.Use(Decompress)
	return e
}
