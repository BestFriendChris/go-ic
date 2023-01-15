package ic

import (
	"bufio"
	"bytes"
	"flag"
	"github.com/BestFriendChris/go-ic/ic/internal/infra/atomic_bool"
	"os"
	"runtime"
	"strings"
)

var (
	updateEnabled *bool
)

func init() {
	updateEnabled = flag.Bool("test.icupdate", false, "allow IC to update test files")
}

func NewTestFileUpdater() DefaultTestFileUpdater {
	return DefaultTestFileUpdater{
		alreadySeen: atomic_bool.NewGlobal(),
	}
}

type DefaultTestFileUpdater struct {
	alreadySeen *atomic_bool.AtomicBool
}

func (d DefaultTestFileUpdater) UpdateEnabled() bool {
	_, envUpdateEnabled := os.LookupEnv("IC_UPDATE")
	return *updateEnabled || envUpdateEnabled
}

func (d DefaultTestFileUpdater) Update(ic *IC, got string) {
	ic.t.Helper()

	_, fName, lineNo, ok := runtime.Caller(3)
	if !ok {
		panic("update was called incorrectly")
	}

	if d.alreadySeen.Set() {
		ic.t.Log(`IC: already updated a test file. Skipping update. Rerun tests to try again`)
		return
	}

	file, err := os.OpenFile(fName, os.O_RDWR, 0644)
	if err != nil {
		ic.t.Log("error opening test file for update")
		ic.t.FailNow()
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)

	var sb bytes.Buffer
	// Skip previous lines
	for i := 0; i < lineNo-1; i++ {
		scanner.Scan()
		sb.WriteString(scanner.Text() + "\n")
	}

	var line string
	for scanner.Scan() {
		line = scanner.Text()
		if strings.IndexAny(line, "`\"") >= 0 {
			// This is the line with the `` or ""
			break
		}
		sb.WriteString(line + "\n")
	}

	idx := strings.Index(line, "Expect")
	// This is the line with Expect on it
	if idx > -1 {
		// We need to write everything up to the '('
		sb.WriteString(line[:idx])
		line = line[idx:]

		idx = strings.Index(line, "(") + 1
		sb.WriteString(line[:idx])
		line = line[idx:]
	}

	idx = strings.IndexAny(line, "`\"")
	if idx == -1 || line[idx] != line[idx+1] {
		panic("should not be possible to not find either `` or \"\"")
	}
	if idx > 0 {
		// There is (most likely) whitespace before the empty quotes
		sb.WriteString(line[:idx])
		line = line[idx:]
	}

	sb.WriteString("`")
	if strings.Index(got, "\n") >= 0 {
		// update as multiline
		sb.WriteString("\n")
	}
	sb.WriteString(got)
	sb.WriteString("`")

	// write out the rest of the line skipping the empty quotes
	sb.WriteString(line[2:] + "\n")

	// write out the rest of the lines
	for scanner.Scan() {
		sb.WriteString(scanner.Text() + "\n")
	}

	ic.t.Log(`IC: Updating test file. Rerun tests to verify`)

	// update the test file!
	_, _ = file.WriteAt(sb.Bytes(), 0)
}
