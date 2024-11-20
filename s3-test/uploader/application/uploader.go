package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	_ "net/http/pprof"

	storagepb "github.com/Bronn1/s3-test/proto"
	meta "github.com/Bronn1/s3-test/uploader/domain"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

const GB = 1 << 30 // 1 GiB
const MaxFileSizeGb = 10 * GB

type UploaderService struct {
	StorageConnections StorageServerConnections
	log                *zap.SugaredLogger
	metaService        meta.FileMetaService
}

type ObjectLocation struct {
	ServerAddr string
	ObjectNum  int
}

func (s *UploaderService) HandleGetFile(w http.ResponseWriter, r *http.Request) {
	if !CheckAuth(r) {
		http.Error(w, "Not authorized", http.StatusForbidden)
		return
	}
	fileId := r.URL.Query().Get("file_id")
	fileMeta, err := s.metaService.GetFileMeta(fileId)
	if err != nil {
		http.Error(w, "Invalid file id", http.StatusNotFound)
		return
	}

	errGroup, _ := errgroup.WithContext(context.Background())
	objectsBuf := make([][]byte, meta.DefaultSplitFileIntoObjects)
	mutex := &sync.Mutex{}
	for i, addr := range fileMeta.SavedOn {
		i := i // for lower than go1.22
		buf := make([]byte, 0)
		errGroup.Go(func() error {
			connection := s.StorageConnections.GetConnectionByAdrr(addr)
			if connection == nil {
				return xerrors.Errorf("Cannot get file from server: %s", addr)
			}
			stream, err := connection.grpcClient.Get(context.Background(), &storagepb.ObjectRequest{FileId: fileId, ObjectNum: int32(i)})
			defer stream.CloseSend()
			if err != nil {
				return err
			}
			for {
				rawObject, err := stream.Recv()

				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}
				chunk := rawObject.GetChunk()
				buf = append(buf, chunk...)

			}
			mutex.Lock()
			defer mutex.Unlock()
			objectsBuf[i] = buf

			return nil
		})
	}
	if err := errGroup.Wait(); err != nil {
		s.log.Errorf("Failed to get file: %v", err.Error())
		http.Error(w, "Failed to get file ", http.StatusInternalServerError)
		return
	}
	res := make([]byte, 0)
	for _, object := range objectsBuf {
		res = append(res, object...)
	}
	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Disposition", "attachment; filename="+fileId)
	w.Write(res)
	s.log.Infof("File: %s with size: %d has been sent to user", fileId, len(res))
}

func (s *UploaderService) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if !CheckAuth(r) {
		http.Error(w, "Not authorized", http.StatusForbidden)
		return
	}
	if err := r.ParseMultipartForm(MaxFileSizeGb); err != nil {
		http.Error(w, fmt.Sprintf("Cannot upload file larger than: %d", MaxFileSizeGb), http.StatusBadRequest)
		return
	}
	file, originalName, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
	}

	defer file.Close()
	errGroup, _ := errgroup.WithContext(context.Background())
	metadata := meta.NewFileMetadata(originalName.Size)
	objectSize := metadata.GetObjectSize()
	usedServersToSaveCh := make(chan ObjectLocation, meta.DefaultSplitFileIntoObjects)

	for i := 0; i < meta.DefaultSplitFileIntoObjects; i++ {
		i := i // for lower than go1.22
		errGroup.Go(func() error {

			buf := make([]byte, objectSize)
			connection := s.StorageConnections.GetFreeConnection()
			stream, err := connection.grpcClient.Upload(context.Background())
			if err != nil {
				return err
			}
			fileOffset := i * int(objectSize)
			num, err := file.ReadAt(buf, int64(fileOffset))
			if err != nil && err != io.EOF {
				return err
			}
			chunk := buf[:num]

			err = stream.Send(&storagepb.ObjectInfo{FileId: metadata.Id.String(), ObjectNum: int32(i), Chunk: chunk})
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			s.log.Infof("Sent - file chunk #%d - size - %v\n", i, len(chunk))

			_, err = stream.CloseAndRecv()
			if err != nil {
				return err
			}
			usedServersToSaveCh <- ObjectLocation{ObjectNum: i, ServerAddr: connection.serverAddr}
			return nil
		})
	}
	if err := errGroup.Wait(); err != nil {
		s.log.Errorf("Failed to upload file: %v", err.Error())
		http.Error(w, "Failed to upload file ", http.StatusInternalServerError)
		return
	}

	close(usedServersToSaveCh)
	for object := range usedServersToSaveCh {
		metadata.SavedOn[object.ObjectNum] = object.ServerAddr
	}
	if err := s.metaService.SaveFileMeta(metadata); err != nil {
		s.log.Errorf("Failed to save file meta to DB: %v", err.Error())
		http.Error(w, "Failed to upload file ", http.StatusInternalServerError)
		return
	}
	s.log.Infof("File: %s saved, original size: %d", metadata.Id, originalName.Size)
	s.WriteUploadResponse(metadata, w)
}

func (s *UploaderService) WriteUploadResponse(fileMeta meta.FileMetadata, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	jsonResponse := make(map[string]string)
	jsonResponse["id"] = fileMeta.Id.String()
	jsonResponse["status"] = "ok"
	json.NewEncoder(w).Encode(jsonResponse)
}
