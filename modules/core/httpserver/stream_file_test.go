package httpserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
)

func newStreamContext() (*echo.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/file", nil)
	return echo.New().NewContext(request, recorder), recorder
}

func TestRangeNotSatisfiable_Answers416WithTheSize(t *testing.T) {
	echoCtx, recorder := newStreamContext()

	assert.NoError(t, rangeNotSatisfiable(echoCtx, 1234))
	assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, recorder.Code)
	assert.Equal(t, "bytes */1234", recorder.Header().Get("Content-Range"))
}

// The domain wraps the storage error; ServeStreamFile must still recognise it.
func TestRangeNotSatisfiable_SurvivesWrapping(t *testing.T) {
	wrapped := errors.Wrap(filestorage.ErrRangeNotSatisfiable{Size: 9}, "open stored object")

	var unsatisfiable filestorage.ErrRangeNotSatisfiable
	assert.True(t, errors.As(wrapped, &unsatisfiable))
	assert.Equal(t, int64(9), unsatisfiable.Size)
}

func TestWriteFileStream_Status(t *testing.T) {
	cases := []struct {
		name         string
		rangeHeader  string
		contentRange *string
		wantStatus   int
	}{
		{"no range", "", nil, http.StatusOK},
		{"range served", "bytes=0-1", util.ToPtr("bytes 0-1/5"), http.StatusPartialContent},
		{"range ignored by storage", "bytes=abc", nil, http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			echoCtx, recorder := newStreamContext()
			request := FileStreamRequestBase{RangeHeader: tc.rangeHeader}
			stream := GetFileStreamResponse{
				Name: "a.txt", MimeType: "text/plain", ContentRange: tc.contentRange,
				Body: io.NopCloser(strings.NewReader("hello")),
			}

			assert.NoError(t, writeFileStream(echoCtx, request, stream, "file"))
			assert.Equal(t, tc.wantStatus, recorder.Code)
		})
	}
}
