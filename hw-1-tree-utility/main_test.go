package main

import (
	"bytes"
	"testing"
)

const testFullResult = `├───project
│	├───file.txt (19b)
│	└───gopher.png (70372b)
├───static
│	├───a_lorem
│	│	├───dolor.txt (empty)
│	│	├───gopher.png (70372b)
│	│	└───ipsum
│	│		└───gopher.png (70372b)
│	├───css
│	│	└───body.css (28b)
│	├───empty.txt (empty)
│	├───html
│	│	└───index.html (57b)
│	├───js
│	│	└───site.js (10b)
│	└───z_lorem
│		├───dolor.txt (empty)
│		├───gopher.png (70372b)
│		└───ipsum
│			└───gopher.png (70372b)
├───zline
│	├───empty.txt (empty)
│	└───lorem
│		├───dolor.txt (empty)
│		├───gopher.png (70372b)
│		└───ipsum
│			└───gopher.png (70372b)
└───zzfile.txt (empty)
`

func TestTreeFull(t *testing.T) {
	out := new(bytes.Buffer)
	dirTree(out, "testdata", true)
	result := out.String()

	if result != testFullResult {
		t.Errorf("test for OK Failed - results not match\nGot:\n%v\nExpected:\n%v", result, testFullResult)
	}
}

const testDirResult = `├───project
├───static
│	├───a_lorem
│	│	└───ipsum
│	├───css
│	├───html
│	├───js
│	└───z_lorem
│		└───ipsum
└───zline
	└───lorem
		└───ipsum
`

func TestTreeDir(t *testing.T) {
	out := new(bytes.Buffer)
	dirTree(out, "testdata", false)
	result := out.String()
	if result != testDirResult {
		t.Errorf("test for OK Failed - results not match\nGot:\n%v\nExpected:\n%v", result, testDirResult)
	}
}

const testdata2DirResult = `├───1_test.go (empty)
├───a_test.go (empty)
├───dir1
│	└───dir2
│		└───dir3
│			└───dir4
│				└───r.txt (2b)
├───empty_dir
└───testdata
	├───project
	│	├───file.txt (19b)
	│	└───gopher.png (70372b)
	├───static
	│	├───a_lorem
	│	│	├───dolor.txt (empty)
	│	│	├───gopher.png (70372b)
	│	│	└───ipsum
	│	│		└───gopher.png (70372b)
	│	├───css
	│	│	└───body.css (28b)
	│	├───empty.txt (empty)
	│	├───html
	│	│	└───index.html (57b)
	│	├───js
	│	│	└───site.js (10b)
	│	└───z_lorem
	│		├───dolor.txt (empty)
	│		├───gopher.png (70372b)
	│		└───ipsum
	│			└───gopher.png (70372b)
	├───zline
	│	├───empty.txt (empty)
	│	└───lorem
	│		├───dolor.txt (empty)
	│		├───gopher.png (70372b)
	│		└───ipsum
	│			└───gopher.png (70372b)
	└───zzfile.txt (empty)
`

func TestData2Fulldir(t *testing.T) {
	out := new(bytes.Buffer)
	dirTree(out, "testdata2", true)
	result := out.String()
	if result != testdata2DirResult {
		t.Errorf("test for OK Failed - results not match\nGot:\n%v\nExpected:\n%v", result, testdata2DirResult)
	}
}

const testEmptyDirResult = ``

func TestTreeEmptyDir(t *testing.T) {
	out := new(bytes.Buffer)
	dirTree(out, "testdata2/empty_dir", true)
	result := out.String()
	if result != testEmptyDirResult {
		t.Errorf("test for OK Failed - results not match\nGot:\n%v\nExpected:\n%v", result, testEmptyDirResult)
	}
}

func TestTreeEmptyDirNoFiles(t *testing.T) {
	out := new(bytes.Buffer)
	dirTree(out, "testdata/static/css", false) // folder contains one file, shouldn't be printed
	result := out.String()

	if result != testEmptyDirResult {
		t.Errorf("test for OK Failed - results not match\nGot:\n%v\nExpected:\n%v", result, testEmptyDirResult)
	}
}

const testDirOneFileResult = `└───site.js (10b)
`

func TestDirWithOneFile(t *testing.T) {
	out := new(bytes.Buffer)
	dirTree(out, "testdata/static/js", true)
	result := out.String()

	if result != testDirOneFileResult {
		t.Errorf("test for OK Failed - results not match\nGot:\n%v\nExpected:\n%v", result, testDirOneFileResult)
	}
}

const testDirOneDirResult = `└───ipsum
`

func TestDirWithOneDir(t *testing.T) {
	out := new(bytes.Buffer)
	dirTree(out, "testdata/static/a_lorem", false)
	result := out.String()

	if result != testDirOneDirResult {
		t.Errorf("test for OK Failed - results not match\nGot:\n%v\nExpected:\n%v", result, testDirOneDirResult)
	}
}

const invalidDirResult = `Error occured: open invaliddir: no such file or directory`

func TestPanicOnInvalidDir(t *testing.T) {
	out := new(bytes.Buffer)
	dirTree(out, "invaliddir", false)
	result := out.String()

	if result != invalidDirResult {
		t.Errorf("test for OK Failed - results not match\nGot:\n%v\nExpected:\n%v", result, invalidDirResult)
	}
}

const notDirErrorResult = `Error occured: readdirent testdata/static/js/site.js: not a directory`

func TestPanicOnFile(t *testing.T) {
	out := new(bytes.Buffer)
	dirTree(out, "testdata/static/js/site.js", true)
	result := out.String()

	if result != notDirErrorResult {
		t.Errorf("test for OK Failed - results not match\nGot:\n%v\nExpected:\n%v", result, notDirErrorResult)
	}
}
