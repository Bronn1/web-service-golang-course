package application

import (
	meta "github.com/Bronn1/s3-test/uploader/domain"
	"go.uber.org/zap"
)

func New(options ...func(*UploaderService)) *UploaderService {
	uploader := &UploaderService{}

	for _, opt := range options {
		opt(uploader)
	}

	return uploader
}

func WithLogger() func(*UploaderService) {
	return func(s *UploaderService) {
		s.log = zap.Must(zap.NewProduction()).Sugar()
	}
}

func WithStorageServers(connections StorageServerConnections) func(*UploaderService) {
	return func(s *UploaderService) {
		s.StorageConnections = connections
	}
}

func WithFileMetaService(fileMeta meta.FileMetaService) func(*UploaderService) {
	return func(s *UploaderService) {
		s.metaService = fileMeta
	}
}
