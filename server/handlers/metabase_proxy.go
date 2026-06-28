package handlers

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func MetabaseProxy(metabaseInternalURL string) gin.HandlerFunc {
	target, _ := url.Parse(metabaseInternalURL)
	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.ModifyResponse = func(resp *http.Response) error {
		ct := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "text/html") {
			return nil
		}

		var reader io.ReadCloser
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			var err error
			reader, err = gzip.NewReader(resp.Body)
			if err != nil {
				return err
			}
			defer reader.Close()
		default:
			reader = resp.Body
		}

		body, err := io.ReadAll(reader)
		resp.Body.Close()
		if err != nil {
			return err
		}

		// Metabase uses relative paths like href="app/..." and src="app/..."
		// which resolve wrong under /metabase/public/dashboard/. Rewrite to absolute.
		body = bytes.ReplaceAll(body, []byte(`href="app/`), []byte(`href="/metabase/app/`))
		body = bytes.ReplaceAll(body, []byte(`src="app/`), []byte(`src="/metabase/app/`))
		body = bytes.ReplaceAll(body, []byte(`href="/"`), []byte(`href="/metabase/"`))

		resp.Body = io.NopCloser(bytes.NewReader(body))
		resp.ContentLength = int64(len(body))
		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
		resp.Header.Del("Content-Encoding")

		// Relax CSP to allow Metabase's CSS-in-JS (Emotion) inline styles
		csp := resp.Header.Get("Content-Security-Policy")
		if csp != "" {
			csp = strings.ReplaceAll(csp, "style-src-attr 'self'", "style-src-attr 'self' 'unsafe-inline'")
			csp = strings.ReplaceAll(csp, "style-src 'self'", "style-src 'self' 'unsafe-inline'")
			resp.Header.Set("Content-Security-Policy", csp)
		}

		return nil
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		stripped := strings.TrimPrefix(path, "/metabase")

		allowed := strings.HasPrefix(stripped, "/public/dashboard/") ||
			strings.HasPrefix(stripped, "/app/") ||
			strings.HasPrefix(stripped, "/api/") ||
			strings.HasPrefix(stripped, "/public/")

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "not allowed"})
			return
		}

		c.Request.URL.Path = stripped
		c.Request.Host = target.Host
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
