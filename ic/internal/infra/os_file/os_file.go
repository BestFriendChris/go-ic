package os_file

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

type OsFileManager struct {
	fo fileOpener
}

func New() *OsFileManager {
	return &OsFileManager{
		fo: &osFileOpener{},
	}
}

func NewNullable(testFiles *map[string]string) *OsFileManager {
	return &OsFileManager{
		fo: &fakeFileOpener{
			fakeFs: testFiles,
		},
	}
}

func (fm OsFileManager) OpenRW(fName string) (*OsFile, error) {
	return fm.fo.OpenFile(fName, os.O_RDWR, 0644)
}

type OsFile struct {
	f fileRewriter
}

func (osf *OsFile) Close() {
	_ = osf.f.Close()
}

func (osf *OsFile) Scanner() *bufio.Scanner {
	return bufio.NewScanner(osf.f)
}

func (osf *OsFile) Rewrite(bytes []byte) error {
	_, err := osf.f.WriteAt(bytes, 0)
	return err
}

/********************************************************************************
private nullable interfaces - fileOpener
********************************************************************************/

type fileOpener interface {
	OpenFile(name string, flag int, perm os.FileMode) (*OsFile, error)
}

type osFileOpener struct{}

func (fo *osFileOpener) OpenFile(name string, flag int, perm os.FileMode) (*OsFile, error) {
	file, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return &OsFile{f: file}, nil
}

type fakeFileOpener struct {
	fakeFs *map[string]string
}

func (f *fakeFileOpener) OpenFile(name string, _ int, _ os.FileMode) (*OsFile, error) {
	body, found := (*f.fakeFs)[name]
	if !found {
		return nil, fmt.Errorf("file %q not found", name)
	}

	bodyBytes := []byte(body)
	reader := bytes.NewReader(bodyBytes)
	writer := bytes.Buffer{}

	return &OsFile{f: &fakeFileRewriter{fName: name, fakeFs: f.fakeFs, reader: reader, writer: &writer}}, nil
}

/********************************************************************************
private nullable interfaces - fileWriter
********************************************************************************/

type fileRewriter interface {
	io.Reader
	Close() error
	WriteAt(b []byte, off int64) (int, error)
}

type fakeFileRewriter struct {
	fName  string
	fakeFs *map[string]string
	reader io.Reader
	writer *bytes.Buffer
}

func (ffr *fakeFileRewriter) Read(b []byte) (n int, err error) {
	return ffr.reader.Read(b)
}

func (ffr *fakeFileRewriter) Close() error {
	// Nothing to close
	return nil
}

func (ffr *fakeFileRewriter) WriteAt(b []byte, _ int64) (int, error) {
	n, err := ffr.writer.Write(b)
	s := ffr.writer.String()
	(*ffr.fakeFs)[ffr.fName] = s
	return n, err
}
