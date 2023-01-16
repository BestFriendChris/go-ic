package os_file

import (
	"bytes"
	"os"
	"path"
	"testing"
)

func Test_fullRealUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("uses real file system")
	}

	content := `line 1
line 2
line 3
`
	fPath := makeTempFile(t, "file.txt", content)

	fileManager := New()
	updateFile(t, fileManager, fPath, func(line string) string {
		return "new " + line + "\n"
	})

	gotBytes, _ := os.ReadFile(fPath)

	got := string(gotBytes)
	want := `new line 1
new line 2
new line 3
`
	if got != want {
		t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func Test_nullable(t *testing.T) {
	content := `line 1
line 2
line 3
`
	fPath := "/tmp/file.txt"
	fakeFiles := makeFakeTempFile(fPath, content)

	fileManager := NewNullable(&fakeFiles)
	updateFile(t, fileManager, fPath, func(line string) string {
		return "new " + line + "\n"
	})

	got := fakeFiles[fPath]
	want := `new line 1
new line 2
new line 3
`
	if got != want {
		t.Errorf("\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func makeTempFile(t *testing.T, fName, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	fPath := path.Join(tmpDir, fName)
	err := os.WriteFile(fPath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	return fPath
}

func makeFakeTempFile(fPath, content string) map[string]string {
	return map[string]string{
		fPath: content,
	}
}

func updateFile(t *testing.T, fileManager *OsFileManager, fPath string, updater func(string) string) {
	{
		file, err := fileManager.OpenRW(fPath)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		scanner := file.Scanner()

		var sb bytes.Buffer

		for scanner.Scan() {
			text := scanner.Text()
			sb.WriteString(updater(text))
		}

		err = file.Rewrite(sb.Bytes())
		if err != nil {
			t.Fatal(err)
		}
	}
}
