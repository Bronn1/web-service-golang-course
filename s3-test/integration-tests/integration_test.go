package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	cfg "github.com/Bronn1/s3-test/config"
	storage "github.com/Bronn1/s3-test/object-storage/application"
	storagepb "github.com/Bronn1/s3-test/proto"
	uploader "github.com/Bronn1/s3-test/uploader/application"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// HERE is simple test with upload and get file
// another tests should be:
// 1. Get invalid file id
// 2. Test response if one storage server is down
// 3. Test upload 5 files and get few
// 4. Upload large file
// 5. upload wrong data format
// 6. check auth
func TestUploadAndGetService(t *testing.T) {
	// starting storage 1
	go func() {
		grpcServer := grpc.NewServer()
		port := 50501
		storagepb.RegisterObjectStorageServer(grpcServer, storage.NewServer(fmt.Sprintf("./data_%d", port)))
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			log.Fatal(err.Error())
		}
		grpcServer.Serve(lis)
	}()
	// starting storage 2
	go func() {
		grpcServer := grpc.NewServer()
		port := 50502
		storagepb.RegisterObjectStorageServer(grpcServer, storage.NewServer(fmt.Sprintf("./data_%d", port)))
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			log.Fatal(err.Error())
		}
		grpcServer.Serve(lis)
	}()

	//starting http uploader server
	conf, _ := cfg.NewConfig("integration.conf")
	serverConnections, _ := uploader.NewSimpleBalancerRoundR(conf.Storages)
	ser := uploader.New(
		uploader.WithLogger(),
		uploader.WithFileMetaService(uploader.NewFileMetaServiceWithMock("./mock_db.json")),
		uploader.WithStorageServers(serverConnections),
	)
	http.HandleFunc("/upload", ser.HandleUpload)
	http.HandleFunc("/get", ser.HandleGetFile)
	go func() {
		if err := http.ListenAndServe(":8082", nil); err != nil {
			log.Fatal(err.Error())
		}
	}()
	// a bit weird but for tests is kinda ok
	time.Sleep(2 * time.Second)
	uploadedId := ""
	// TEST Creating file and upload to server
	t.Run("Upload file", func(t *testing.T) {
		// create test files
		fileDir, _ := os.Getwd()
		fileName := "upload-file.txt"
		filePath := path.Join(fileDir, fileName)
		createFile, err := os.Create(filePath)
		createFile.Write([]byte("test message in file?&!&&@&!"))
		createFile.Close()
		assert.NoError(t, err)
		file, _ := os.Open(filePath)

		defer file.Close()
		// send file
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
		io.Copy(part, file)
		writer.Close()

		r, err := http.NewRequest("POST", "http://127.0.0.1:8082/upload", body)
		r.Header.Add("Content-Type", writer.FormDataContentType())
		assert.NoError(t, err)

		resp, err := http.DefaultClient.Do(r)
		assert.NoError(t, err)
		response := make(map[string]string, 0)
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		uploadedId = response["id"]

		req, err := http.NewRequest("GET", "http://localhost:8082/get?file_id="+uploadedId, nil)
		assert.NoError(t, err)

		client := &http.Client{}
		getResp, err := client.Do(req)
		assert.NoError(t, err)

		// Проверяем, что файл был восстановлен корректно
		bodyGet, err := io.ReadAll(getResp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "test message in file?&!&&@&!", string(bodyGet))
	})

}
