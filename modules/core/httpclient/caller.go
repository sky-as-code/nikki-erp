package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sky-as-code/nikki-erp/modules/core/httpclient/client"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"go.bryk.io/pkg/errors"
)

const (
	defaultMaxResponseBytes = 10 << 20
	errBodySnippetMax       = 2048
)

type HttpCaller struct {
	baseUrl string
	client  *client.HttpClient
	logger  logging.LoggerService
}

type Request struct {
	Method  string      `json:"method"`
	Path    string      `json:"path"`
	Headers http.Header `json:"header"`
	Query   url.Values  `json:"query"`
	Body    any         `json:"body"`
}

type Response struct {
	StatusCode int         `json:"status_code"`
	Headers    http.Header `json:"headers"`
	Body       []byte      `json:"body"`
}

func NewHttpCaller(baseUrl string, client *client.HttpClient, logger logging.LoggerService) *HttpCaller {
	trimmed := strings.TrimSpace(baseUrl)
	if trimmed == "" {
		panic("base url is required")
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		panic(err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		panic("base url scheme must be http or https")
	}

	if u.Host == "" {
		panic("base url must include host")
	}

	normalized := strings.TrimRight(u.String(), "/")
	return &HttpCaller{client: client, baseUrl: normalized, logger: logger}
}

func joinBasePathQuery(baseUrl, path string, query url.Values) (string, error) {
	elem := strings.TrimPrefix(strings.TrimSpace(path), "/")
	fullURL, err := url.JoinPath(baseUrl, elem)
	if err != nil {
		return "", errors.Wrap(err, "join url path")
	}

	if len(query) == 0 {
		return fullURL, nil
	}

	u, err := url.Parse(fullURL)
	if err != nil {
		return "", errors.Wrap(err, "parse joined url")
	}

	u.RawQuery = query.Encode()
	return u.String(), nil
}

func requestBodyReader(body any) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, errors.Wrap(err, "marshal request body")
	}

	return bytes.NewReader(b), nil
}

func errBodySnippet(b []byte) string {
	if len(b) <= errBodySnippetMax {
		return string(b)
	}
	return string(b[:errBodySnippetMax]) + "..."
}

func (this *HttpCaller) exec(ctx context.Context, req *Request, acceptJSON bool) (int, http.Header, []byte, error) {
	fullURL, err := joinBasePathQuery(this.baseUrl, req.Path, req.Query)
	if err != nil {
		return 0, nil, nil, err
	}

	bodyReader, err := requestBodyReader(req.Body)
	if err != nil {
		return 0, nil, nil, err
	}

	method := req.Method
	if method == "" {
		method = http.MethodGet
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return 0, nil, nil, errors.Wrap(err, "build http request")
	}

	if req.Headers != nil {
		for k, v := range req.Headers {
			for _, vv := range v {
				httpReq.Header.Add(k, vv)
			}
		}
	}

	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	if acceptJSON && httpReq.Header.Get("Accept") == "" {
		httpReq.Header.Set("Accept", "application/json")
	}

	httpResp, err := this.client.Do(httpReq)
	if err != nil {
		return 0, nil, nil, errors.Wrap(err, "http round trip")
	}

	defer httpResp.Body.Close()
	limited := http.MaxBytesReader(nil, httpResp.Body, defaultMaxResponseBytes)
	respBody, err := io.ReadAll(limited)
	if err != nil {
		return 0, nil, nil, errors.Wrap(err, "read response body")
	}

	return httpResp.StatusCode, httpResp.Header, respBody, nil
}

func (this *HttpCaller) Do(ctx context.Context, req *Request) (*Response, error) {
	status, hdr, respBody, err := this.exec(ctx, req, true)
	if err != nil {
		return nil, err
	}

	out := &Response{StatusCode: status, Headers: hdr, Body: respBody}

	this.LogRequest(req, out)

	if status < 200 || status > 299 {
		return out, errors.Errorf("http error: status %d, body: %s", status, errBodySnippet(respBody))
	}

	return out, nil
}

func (this *HttpCaller) LogRequest(req *Request, res *Response) {
	this.logger.Info("[HTTP Caller] Send request", logging.Attr{
		"method":          req.Method,
		"path":            req.Path,
		"request_headers": req.Headers,
		"request_query":   req.Query,
		"status_code":     res.StatusCode,
		"res_headers":     res.Headers,
		"res_body":        string(res.Body),
	})
}
