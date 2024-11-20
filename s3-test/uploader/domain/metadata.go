package domain

import (
	guuid "github.com/google/uuid"
	"golang.org/x/xerrors"
)

const (
	DefaultSplitFileIntoObjects = 6
)

type FileMetaService struct {
	MetaDbRepo FileMetaDbRepo
}

func (m *FileMetaService) SaveFileMeta(metadata FileMetadata) error {
	return m.MetaDbRepo.CreateOrUpdateFile(metadata)
}

func (m *FileMetaService) GetFileMeta(fileIdStr string) (*FileMetadata, error) {
	fileId, err := guuid.Parse(fileIdStr)
	if err != nil {
		return nil, xerrors.Errorf("Invalid fail id: %s", err.Error())
	}
	fileMeta := m.MetaDbRepo.GetMetadata(fileId)
	if fileMeta == nil {
		return nil, xerrors.Errorf("Missing file meta: %s", fileId.String())
	}
	return fileMeta, nil
}

type Login string

type FileMetadata struct {
	Id         guuid.UUID `json:"id"`
	SavedOn    []string   `json:"savedOn"`
	Md5Sum     string     `json:"md5"`
	Owner      Login      `json:"owner"`
	Size       int64      `json:"size"`
	BucketName string     // TODO unused for now. To simplify things we gonna generate just unique ids for files but ideally user should have buckets and filename
	Filetype   string
}

func NewFileMetadata(filesize int64) FileMetadata {
	return FileMetadata{Id: guuid.New(), Size: filesize, SavedOn: make([]string, DefaultSplitFileIntoObjects)}
}
func (fm *FileMetadata) GetObjectSize() int64 {
	return (fm.Size / DefaultSplitFileIntoObjects) + 5
}

type FileMetaDbRepo interface {
	CreateOrUpdateFile(fileMeta FileMetadata) error
	GetMetadata(fileId guuid.UUID) *FileMetadata
}
