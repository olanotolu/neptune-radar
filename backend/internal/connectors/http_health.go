package connectors

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"time"
)

// HTTPHealthConnector performs a real GET against an endpoint URL, times it,
// and hashes a stable slice of the response body so future page-structure
// drift is detectable across runs. This is the connector for government
// record-search pages and church bulletin archives — reachability and
// structure, not extraction.
type HTTPHealthConnector struct {
	Client *http.Client
}

func NewHTTPHealthConnector() *HTTPHealthConnector {
	return &HTTPHealthConnector{Client: &http.Client{Timeout: 15 * time.Second}}
}

func (c *HTTPHealthConnector) CheckHealth(ctx context.Context, endpointURL string) CheckResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL, nil)
	if err != nil {
		return CheckResult{Status: "failure", ErrorMessage: err.Error()}
	}
	req.Header.Set("User-Agent", "NeptuneRadar-SourceHealthCheck/1.0 (+https://meetneptune.com; source registry health probe)")

	start := time.Now()
	resp, err := c.Client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		return CheckResult{Status: "failure", ResponseTimeMs: int(elapsed.Milliseconds()), ErrorMessage: err.Error()}
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	sig := structureSignature(body)

	if resp.StatusCode >= 400 {
		return CheckResult{
			Status: "failure", HTTPStatus: resp.StatusCode, ResponseTimeMs: int(elapsed.Milliseconds()),
			StructureSignature: sig, ErrorMessage: fmt.Sprintf("HTTP %d", resp.StatusCode),
		}
	}
	if readErr != nil {
		return CheckResult{Status: "failure", HTTPStatus: resp.StatusCode, ResponseTimeMs: int(elapsed.Milliseconds()), ErrorMessage: readErr.Error()}
	}
	return CheckResult{
		Status: "success", HTTPStatus: resp.StatusCode, ResponseTimeMs: int(elapsed.Milliseconds()),
		StructureSignature: sig,
	}
}

func structureSignature(body []byte) string {
	h := fnv.New64a()
	_, _ = h.Write(body)
	return fmt.Sprintf("%x", h.Sum64())
}
