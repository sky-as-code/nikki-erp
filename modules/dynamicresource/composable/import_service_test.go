package composable

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
)

// fakeStorage keeps objects in memory and records the lifecycle the import must follow.
type fakeStorage struct {
	objects map[string][]byte
	puts    []string
	removed []string
	failPut bool
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{objects: map[string][]byte{}}
}

func (this *fakeStorage) Put(_ context.Context, key string, r io.Reader, _ *filestorage.PutOptions) error {
	if this.failPut {
		return assert.AnError
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	this.objects[key] = data
	this.puts = append(this.puts, key)
	return nil
}

func (this *fakeStorage) Open(_ context.Context, key string, _ string) (*filestorage.StreamObjectResult, error) {
	data, ok := this.objects[key]
	if !ok {
		return nil, assert.AnError
	}
	return &filestorage.StreamObjectResult{Body: io.NopCloser(bytes.NewReader(data))}, nil
}

func (this *fakeStorage) Remove(_ context.Context, key string) error {
	delete(this.objects, key)
	this.removed = append(this.removed, key)
	return nil
}

func (this *fakeStorage) GeneratePresignedUrl(context.Context, string, time.Duration) (string, error) {
	return "", nil
}

// multipartFile builds a real *multipart.FileHeader the way a browser upload arrives.
func multipartFile(t *testing.T, name string, content string) *multipart.FileHeader {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(ImportFormFile, name)
	require.NoError(t, err)
	_, err = io.WriteString(part, content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	form, err := multipart.NewReader(body, writer.Boundary()).ReadForm(maxImportFormMemory)
	require.NoError(t, err)
	return form.File[ImportFormFile][0]
}

// importFixture wires the default application service over the import schemas with a fake
// storage and a fake category onion, the way the real onion would at boot.
type importFixture struct {
	appSvc   CrudApplicationService
	domSvc   *fakeDomainService
	category *fakeOnion
	storage  *fakeStorage
}

func newImportFixture(t *testing.T) *importFixture {
	t.Helper()
	owner, category := newImportSchemas(t)
	categoryOnion := &fakeOnion{domSvc: newFakeDomainService(category)}
	categoryOnion.domSvc.idPrefix = "cat_"
	registerSource(category.Name(), categoryOnion.domSvc.repo, categoryOnion)

	storage := newFakeStorage()
	domSvc, appSvc := newAppService(owner, NewAppServiceParam{
		Storage:      storage,
		ImportLimits: ImportLimits{MaxFileSizeBytes: 1 << 20, AllowedExtensions: []string{"csv", "xlsx"}, MaxRows: 100},
	})
	domSvc.idPrefix = "prod_"
	return &importFixture{appSvc: appSvc, domSvc: domSvc, category: categoryOnion, storage: storage}
}

const importCsv = "ID,Tên,SL,Phân loại\nE1,Coke,3,Beverage\nE2,Pepsi,x,Beverage\nE3,Chips,1,Snack\n"

func importCommand(header *multipart.FileHeader, createMissing bool) ImportCommand {
	return ImportCommand{
		Params:     dmodel.DynamicFields{basemodel.FieldOrgId: "org_mine"},
		FileHeader: header,
		Mapping: ImportMapping{
			LanguageCode:            "vi-VN",
			CreateMissingReferences: createMissing,
			Columns: []ColumnMapping{
				{Source: "ID", Target: FieldExternalId}, {Source: "Tên", Target: "name"},
				{Source: "SL", Target: "qty"}, {Source: "Phân loại", Target: "category_id"},
			},
		},
	}
}

func TestImportEndToEndWritesRowsAndCleansUpStorage(t *testing.T) {
	fx := newImportFixture(t)
	fx.category.domSvc.repo.(*stubRepository).searchItems = []dmodel.DynamicFields{
		{"id": "c1", "name": "Beverage", basemodel.FieldOrgId: "org_mine"},
	}

	result, err := fx.appSvc.Import(memberContext(), importCommand(multipartFile(t, "products.csv", importCsv), false))

	require.NoError(t, err)
	require.Equal(t, 0, result.ClientErrors.Count(), result.ClientErrors)
	assert.Equal(t, 3, result.Data.TotalRows)
	assert.Equal(t, 1, result.Data.CreatedCount)
	require.Len(t, result.Data.Errors, 2)
	assert.Equal(t, ErrRowCellInvalid, result.Data.Errors[0].Code)
	assert.Equal(t, 2, result.Data.Errors[0].Row)
	assert.Equal(t, ErrRowReferenceNotFound, result.Data.Errors[1].Code)
	assert.Equal(t, 3, result.Data.Errors[1].Row)

	require.Len(t, fx.domSvc.history, 1)
	written := fx.domSvc.history[0]
	assert.Equal(t, "c1", written["category_id"])
	assert.Equal(t, "org_mine", written[basemodel.FieldOrgId])
	assert.Equal(t, SourceSystemImport, written[FieldSourceSystem])
	assert.Equal(t, int32(3), written["qty"])

	require.Len(t, fx.storage.puts, 1)
	assert.True(t, strings.HasPrefix(fx.storage.puts[0], "imports/"))
	assert.True(t, strings.HasSuffix(fx.storage.puts[0], ".csv"))
	assert.Equal(t, fx.storage.puts, fx.storage.removed, "the object is removed after processing")
	assert.Empty(t, fx.storage.objects)
}

func TestImportCreatesMissingReferencesWhenAsked(t *testing.T) {
	fx := newImportFixture(t)

	result, err := fx.appSvc.Import(memberContext(), importCommand(multipartFile(t, "p.csv", importCsv), true))

	require.NoError(t, err)
	require.Equal(t, 0, result.ClientErrors.Count(), result.ClientErrors)
	assert.Equal(t, 2, result.Data.CreatedCount)
	assert.Len(t, fx.category.domSvc.history, 2, "Beverage and Snack are created once each")
}

func TestImportRefusesWithoutPermissionBeforeTouchingStorage(t *testing.T) {
	fx := newImportFixture(t)

	result, err := fx.appSvc.Import(deniedContext(), importCommand(multipartFile(t, "p.csv", importCsv), false))

	require.NoError(t, err)
	assert.Greater(t, result.ClientErrors.Count(), 0)
	assert.Empty(t, fx.storage.puts)
}

func TestImportRefusesAWrongExtensionAndAMissingFile(t *testing.T) {
	fx := newImportFixture(t)

	result, err := fx.appSvc.Import(memberContext(), importCommand(multipartFile(t, "p.xls", importCsv), false))
	require.NoError(t, err)
	assert.Equal(t, "err_file_type_not_allowed", result.ClientErrors[0].Key)

	result, err = fx.appSvc.Import(memberContext(), importCommand(nil, false))
	require.NoError(t, err)
	assert.Equal(t, ErrImportFileRequired, result.ClientErrors[0].Key)
	assert.Empty(t, fx.storage.puts)
}

func TestImportReportsTooManyRowsAndStillRemovesTheObject(t *testing.T) {
	fx := newImportFixture(t)
	fx.appSvc.(*DefaultApplicationServiceImpl).importLimits.MaxRows = 1

	result, err := fx.appSvc.Import(memberContext(), importCommand(multipartFile(t, "p.csv", importCsv), false))

	require.NoError(t, err)
	assert.Equal(t, ErrImportTooManyRows, result.ClientErrors[0].Key)
	assert.Equal(t, fx.storage.puts, fx.storage.removed)
}

func TestImportReportsMappingProblemsAsClientErrors(t *testing.T) {
	fx := newImportFixture(t)
	cmd := importCommand(multipartFile(t, "p.csv", importCsv), false)
	cmd.Mapping.Columns = cmd.Mapping.Columns[1:]

	result, err := fx.appSvc.Import(memberContext(), cmd)

	require.NoError(t, err)
	assert.Equal(t, ErrImportMandatoryUnmapped, result.ClientErrors[0].Key)
	assert.Equal(t, fx.storage.puts, fx.storage.removed)
}

func TestImportWithoutStorageIsAConfigurationRefusal(t *testing.T) {
	owner, _ := newImportSchemas(t)
	_, appSvc := newAppService(owner)

	result, err := appSvc.Import(memberContext(), importCommand(multipartFile(t, "p.csv", importCsv), false))

	require.NoError(t, err)
	assert.Equal(t, ErrImportStorageUnavailable, result.ClientErrors[0].Key)
}

// multipartRequest is a browser-shaped POST {resource}/import.
func multipartRequest(t *testing.T, target string, parts map[string]string, fileName string, content string) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for name, value := range parts {
		require.NoError(t, writer.WriteField(name, value))
	}
	if fileName != "" {
		part, err := writer.CreateFormFile(ImportFormFile, fileName)
		require.NoError(t, err)
		_, err = io.WriteString(part, content)
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	request := httptest.NewRequest(http.MethodPost, target, body)
	request.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	return request
}

func newImportRestEngine(fx *importFixture) RestEngine {
	rest := &CrudRestBase{}
	rest.SetApplicationService(fx.appSvc)
	return NewRestEngine(fx.appSvc.Schema().Name(), rest).AddCrudRoutes(CrudActionImport)
}

const mappingJson = `{"language_code":"vi-VN","columns":[{"source":"ID","target":"external_id"},{"source":"Tên","target":"name"},{"source":"SL","target":"qty"},{"source":"Phân loại","target":"category_id"}]}`

func TestRestImportAnswersTheBulkResultWithOrgFromTheQuery(t *testing.T) {
	fx := newImportFixture(t)
	fx.category.domSvc.repo.(*stubRepository).searchItems = []dmodel.DynamicFields{
		{"id": "c1", "name": "Beverage", basemodel.FieldOrgId: "org_mine"},
	}
	request := multipartRequest(t, "/cmp_imp_product/import?org_id=org_mine",
		map[string]string{ImportFormMapping: mappingJson}, "p.csv", importCsv)

	recorder := serveWithContext(t, newImportRestEngine(fx), memberContext(), request)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	assert.Contains(t, recorder.Body.String(), `"created_count":1`)
	assert.Contains(t, recorder.Body.String(), `"total_rows":3`)
	assert.Equal(t, "org_mine", fx.domSvc.history[0][basemodel.FieldOrgId])
}

func TestRestImportRefusesAnUnknownPartAndABadMapping(t *testing.T) {
	fx := newImportFixture(t)

	extra := multipartRequest(t, "/cmp_imp_product/import?org_id=org_mine",
		map[string]string{ImportFormMapping: mappingJson, "surprise": "1"}, "p.csv", importCsv)
	recorder := serveWithContext(t, newImportRestEngine(fx), memberContext(), extra)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "surprise")

	broken := multipartRequest(t, "/cmp_imp_product/import?org_id=org_mine",
		map[string]string{ImportFormMapping: "{not json"}, "p.csv", importCsv)
	recorder = serveWithContext(t, newImportRestEngine(fx), memberContext(), broken)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), ErrImportMappingInvalid)
	assert.Empty(t, fx.storage.puts)
}

func TestRestImportRefusesAnOversizedFile(t *testing.T) {
	fx := newImportFixture(t)
	fx.appSvc.(*DefaultApplicationServiceImpl).importLimits.MaxFileSizeBytes = 10
	request := multipartRequest(t, "/cmp_imp_product/import?org_id=org_mine",
		map[string]string{ImportFormMapping: mappingJson}, "p.csv", importCsv)

	recorder := serveWithContext(t, newImportRestEngine(fx), memberContext(), request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "err_file_too_large")
}
