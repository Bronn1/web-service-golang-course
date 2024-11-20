package application

import (
	"encoding/json"
	"net/http"
	"os"

	meta "github.com/Bronn1/s3-test/uploader/domain"
	guuid "github.com/google/uuid"
)

type MockDb struct {
	FileMeta map[string]meta.FileMetadata
	path     string
}

func NewMockDb(path string) *MockDb {
	mock := &MockDb{FileMeta: map[string]meta.FileMetadata{}, path: path}
	byteValue, _ := os.ReadFile(path)
	_ = json.Unmarshal(byteValue, &mock.FileMeta)
	return mock
}

func (m *MockDb) CreateOrUpdateFile(fileMeta meta.FileMetadata) error {
	m.FileMeta[fileMeta.Id.String()] = fileMeta
	// rewriting every time, just for tests
	file, _ := json.MarshalIndent(m.FileMeta, "", "\t")
	_ = os.WriteFile(m.path, file, 0644)

	return nil
}

func (m *MockDb) GetMetadata(fileId guuid.UUID) *meta.FileMetadata {
	val, ok := m.FileMeta[fileId.String()]
	if ok {
		return &val
	}
	return nil
}

func NewFileMetaServiceWithMock(pathToMockDB string) meta.FileMetaService {
	return meta.FileMetaService{MetaDbRepo: NewMockDb(pathToMockDB)}
}

func CheckAuth(req *http.Request) bool {
	if auth, err := req.Cookie("s3_token"); err == nil && auth.Value == "not_very_secret_auth" {
		return true
	}

	return true
}
