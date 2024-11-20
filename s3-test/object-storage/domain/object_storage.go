package domain

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/xerrors"
)

type Object struct {
	Num        int32
	FileId     string
	Path       string
	OutputFile *os.File
}

func NewObject() *Object {
	return &Object{}
}

func GetObjectByID(fileId string, workDir string, objectNum int32) (*Object, error) {
	objectPath := filepath.Join(workDir, fileId, fmt.Sprintf("%d", objectNum))
	if _, err := os.Stat(objectPath); err == nil {
		file, err := os.Open(objectPath)
		if err != nil {
			return nil, err
		}

		return &Object{Num: objectNum, FileId: fileId, Path: objectPath, OutputFile: file}, nil
	}

	return nil, xerrors.Errorf("Cannot find file")
}

func (obj *Object) InitFile(objectId string, objectDir string, partNum int32) error {
	obj.FileId = objectId
	obj.Num = partNum
	if err := os.Mkdir(objectDir, os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}

	objectPath := filepath.Join(objectDir, fmt.Sprintf("%v", partNum))
	obj.Path = objectPath
	f, err := os.Create(objectPath)
	if err != nil {
		return err
	}
	obj.OutputFile = f

	return nil
}

func (obj *Object) Remove() error {
	if obj.OutputFile == nil {
		return nil
	}
	if err := obj.Close(); err != nil {
		return err
	}
	err := os.Remove(obj.Path)
	obj.OutputFile = nil
	return err
}

func (obj *Object) IsInialized() bool {
	return obj.OutputFile != nil
}

func (obj *Object) ReadChunk(chunk []byte) (int, error) {
	if obj.OutputFile == nil {
		return 0, xerrors.Errorf("File hasn't been initialized yet")
	}

	return obj.OutputFile.Read(chunk)
}

func (obj *Object) WriteChunk(chunk []byte) error {
	if obj.OutputFile == nil {
		return xerrors.Errorf("File hasn't been initialized yet")
	}

	_, err := obj.OutputFile.Write(chunk)
	if err != nil {
		return err
	}

	return nil
}

func (obj *Object) Close() error {
	if obj.IsInialized() {
		return nil
	}

	return obj.OutputFile.Close()
}
