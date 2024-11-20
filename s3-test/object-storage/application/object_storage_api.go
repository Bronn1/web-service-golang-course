package app

import (
	"io"
	"os"
	"path/filepath"

	storage "github.com/Bronn1/s3-test/object-storage/domain"
	storagepb "github.com/Bronn1/s3-test/proto"
	"go.uber.org/zap"
)

const (
	BufferSize = 1024 * 50
)

type ObjectStorageServer struct {
	storagepb.UnimplementedObjectStorageServer
	log     *zap.SugaredLogger
	WorkDir string
}

func NewServer(workdir string) storagepb.ObjectStorageServer {
	logger := zap.Must(zap.NewProduction())
	if err := os.MkdirAll(workdir, os.ModePerm); err != nil {
		logger.Fatal("Cannot create work dir: " + err.Error())
	}

	return &ObjectStorageServer{WorkDir: workdir, log: logger.Sugar()}
}

func (s *ObjectStorageServer) Get(objectRequest *storagepb.ObjectRequest, streamResp storagepb.ObjectStorage_GetServer) error {
	objectInfo, err := storage.GetObjectByID(objectRequest.FileId, s.WorkDir, objectRequest.ObjectNum)
	if err != nil {
		s.log.Errorf("Cannot find object with id: %s, and object number: %s, %s", objectRequest.FileId, objectRequest.ObjectNum, err.Error())
		return err
	}
	defer func() {
		if err := objectInfo.Close(); err != nil {
			s.log.Error("Cannot close file: " + err.Error())
		}
	}()

	for {
		buf := make([]byte, BufferSize)
		n, err := objectInfo.ReadChunk(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			s.log.Error("Cannot read object: %s", err.Error())
			return err
		}
		if err := streamResp.Send(&storagepb.ObjectInfo{ObjectNum: objectRequest.ObjectNum, FileId: objectRequest.FileId, Chunk: buf[:n]}); err != nil {
			s.log.Error("Cannot send object: %s", err.Error())
			return err
		}
	}
	s.log.Infof("Object num: %d with ID %s has been successfully sent", objectRequest.ObjectNum, objectInfo.FileId)

	return nil
}

func (s *ObjectStorageServer) Upload(stream storagepb.ObjectStorage_UploadServer) error {
	object := storage.NewObject()
	defer func() {
		if err := object.Close(); err != nil {
			s.log.Error("Cannot close file: " + err.Error())
		}
	}()
	objectsize := 0
	for {
		rawObject, err := stream.Recv()
		if !object.IsInialized() {
			object.InitFile(rawObject.GetFileId(), filepath.Join(s.WorkDir, rawObject.GetFileId()), rawObject.GetObjectNum())
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			s.log.Error("Error while downloading object: " + err.Error())
			if err := object.Remove(); err != nil {
				s.log.Error("Cannot remove invalid object: " + err.Error())
			}
			return err
		}
		chunk := rawObject.GetChunk()
		objectsize += len(chunk)
		if err := object.WriteChunk(chunk); err != nil {
			s.log.Error("Cannot write another chunk of data:" + err.Error())
			return err
		}

	}
	s.log.Infof("Object %s part # %d are saved, size: %d", object.FileId, object.Num, objectsize)
	return stream.SendAndClose(&storagepb.ObjectUploadStatus{FileId: object.FileId, ObjectNum: object.Num})
}
