package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.bryk.io/pkg/errors"
)

const (
	defaultHTTPTimeout      = 30 * time.Second
	defaultMaxResponseBytes = 10 << 20
	errBodySnippetMax       = 2048
)

type HTTPClient struct {
	client  *http.Client
	baseURL string
}

type Request struct {
	Method  string
	Path    string
	Headers http.Header
	Query   url.Values
	Body    any
}

type Response[T any] struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Data       T
}

type RawResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func NewHTTPClient(baseURL string, hc *http.Client) (*HTTPClient, error) {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return nil, errors.New("base url is required")
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		return nil, errors.Wrap(err, "parse base url")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("base url scheme must be http or https")
	}

	if u.Host == "" {
		return nil, errors.New("base url must include host")
	}

	if hc == nil {
		hc = &http.Client{Timeout: defaultHTTPTimeout}
	}

	normalized := strings.TrimRight(u.String(), "/")
	return &HTTPClient{client: hc, baseURL: normalized}, nil
}

func joinBasePathQuery(baseURL, path string, query url.Values) (string, error) {
	elem := strings.TrimPrefix(strings.TrimSpace(path), "/")
	fullURL, err := url.JoinPath(baseURL, elem)
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

func (this *HTTPClient) exec(ctx context.Context, req *Request, acceptJSON bool) (int, http.Header, []byte, error) {
	fullURL, err := joinBasePathQuery(this.baseURL, req.Path, req.Query)
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

func (this *HTTPClient) DoRaw(ctx context.Context, req *Request) (*RawResponse, error) {
	status, hdr, body, err := this.exec(ctx, req, false)
	if err != nil {
		return nil, err
	}

	out := &RawResponse{StatusCode: status, Headers: hdr, Body: body}
	if status < 200 || status > 299 {
		return out, errors.Errorf("http error: status %d, body: %s", status, errBodySnippet(body))
	}

	return out, nil
}

func Do[T any](this *HTTPClient, ctx context.Context, req *Request) (*Response[T], error) {
	status, hdr, respBody, err := this.exec(ctx, req, true)
	if err != nil {
		return nil, err
	}

	out := &Response[T]{StatusCode: status, Headers: hdr, Body: respBody}
	if status < 200 || status > 299 {
		return out, errors.Errorf("http error: status %d, body: %s", status, errBodySnippet(respBody))
	}

	var data T
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &data); err != nil {
			return out, errors.Wrap(err, "decode json response")
		}
	}

	out.Data = data
	return out, nil
}
