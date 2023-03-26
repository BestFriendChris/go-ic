package ic

import (
	"bytes"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/BestFriendChris/go-ic/ic/internal/infra/atomic_bool"
	"github.com/BestFriendChris/go-ic/ic/internal/infra/cmd"
	"github.com/BestFriendChris/go-ic/ic/internal/infra/os_file"
)

func NewTestFileUpdater() TestFileUpdater {
	return TestFileUpdater{
		alreadySeen:   atomic_bool.NewGlobal(),
		osFileManager: os_file.New(),
		cmd:           cmd.New(),
	}
}

func NewNullableTestFileUpdater(testFiles *map[string]string) (TestFileUpdater, *atomic.Bool, *cmd.OverridableFlagChecker) {
	alreadySeen, underlyingBool := atomic_bool.NewNullable()
	osFileManager := os_file.NewNullable(testFiles)
	c, ofc := cmd.NewNullable()
	return TestFileUpdater{
		alreadySeen:   alreadySeen,
		osFileManager: osFileManager,
		cmd:           c,
	}, underlyingBool, ofc
}

type TestFileUpdater struct {
	alreadySeen   *atomic_bool.AtomicBool
	osFileManager *os_file.OsFileManager
	cmd           *cmd.Cmd
}

func (d TestFileUpdater) UpdateEnabled() bool {
	return d.cmd.IsUpdateEnabled()
}

func (d TestFileUpdater) Update(ic *IC, got string) {
	ic.t.Helper()

	_, fName, lineNo, ok := runtime.Caller(3)
	if !ok {
		panic("update was called incorrectly")
	}

	if d.alreadySeen.Set() {
		ic.t.Log(`IC: already updated a test file. Skipping update. Rerun tests to try again`)
		return
	}

	osFile, err := d.osFileManager.OpenRW(fName)
	if err != nil {
		ic.t.Log("error opening test file for update")
		ic.t.FailNow()
	}
	defer osFile.Close()

	scanner := osFile.Scanner()

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

	var numTabs int
	idx := strings.Index(line, "Expect(")
	// This is the line with Expect on it
	if idx > -1 {
		numTabs = strings.Count(line[:idx], "\t")
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
	isMultiline := strings.Index(got, "\n") >= 0
	if isMultiline {
		// update as multiline
		sb.WriteString("\n")
	}
	got = strings.ReplaceAll(got, "`", "` + \"`\" + `")
	if isMultiline && numTabs > 0 {
		prefixTabs := strings.Repeat("\t", numTabs+1)
		sb.WriteString(prefixTabs)
		sb.WriteString(strings.ReplaceAll(got, "\n", "\n"+prefixTabs))
	} else {
		sb.WriteString(got)
	}
	sb.WriteString("`")

	// write out the rest of the line skipping the empty quotes
	sb.WriteString(line[2:] + "\n")

	// write out the rest of the lines
	for scanner.Scan() {
		sb.WriteString(scanner.Text() + "\n")
	}

	ic.t.Log(`IC: Updating test file. Rerun tests to verify`)

	// rewrite the test file!
	err = osFile.Rewrite(sb.Bytes())
	if err != nil {
		ic.t.Log("error writing test file on update")
		ic.t.FailNow()
	}
}
