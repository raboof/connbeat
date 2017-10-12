package processes

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/elastic/beats/libbeat/logp"
	"github.com/stvp/assert"
)

func TestCmdline(t *testing.T) {
	procs, err := processes(true, false, "/proc")
	assert.Nil(t, err)
	for _, process := range procs {
		if strings.Contains(process.Cmdline, "processes.test") {
			return
		}
	}
	fmt.Println("No process found with cmdline containing 'processes.test'")
	t.Fail()
}

func TestNoCmdline(t *testing.T) {
	procs, err := processes(false, true, "/proc")
	assert.Nil(t, err)
	for _, process := range procs {
		if process.Cmdline != "" {
			fmt.Println("Process found with cmdline even though we're configured not to expose that")
			t.Fail()
		}
	}
}

type testProcFile struct {
	path     string
	contents string
	isLink   bool
}

func createFakeDirectoryStructure(prefix string, files []testProcFile) error {

	var err error
	for _, file := range files {
		dir := filepath.Dir(file.path)
		err = os.MkdirAll(filepath.Join(prefix, dir), 0755)
		if err != nil {
			return err
		}

		if !file.isLink {
			err = ioutil.WriteFile(filepath.Join(prefix, file.path),
				[]byte(file.contents), 0644)
			if err != nil {
				return err
			}
		} else {
			err = os.Symlink(file.contents, filepath.Join(prefix, file.path))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func TestFindSocketsOfPid(t *testing.T) {
	logp.LogInit(logp.LOG_DEBUG, "", false, true, []string{})

	proc := []testProcFile{
		{path: "/proc/766/fd/0", isLink: true, contents: "/dev/null"},
		{path: "/proc/766/fd/1", isLink: true, contents: "/dev/null"},
		{path: "/proc/766/fd/10", isLink: true, contents: "/var/log/nginx/packetbeat.error.log"},
		{path: "/proc/766/fd/11", isLink: true, contents: "/var/log/nginx/sipscan.access.log"},
		{path: "/proc/766/fd/12", isLink: true, contents: "/var/log/nginx/sipscan.error.log"},
		{path: "/proc/766/fd/13", isLink: true, contents: "/var/log/nginx/localhost.access.log"},
		{path: "/proc/766/fd/14", isLink: true, contents: "socket:[7619]"},
		{path: "/proc/766/fd/15", isLink: true, contents: "socket:[7620]"},
		{path: "/proc/766/fd/5", isLink: true, contents: "/var/log/nginx/access.log"},
	}

	// Create fake proc file system
	pathPrefix, err := ioutil.TempDir("/tmp", "")
	if err != nil {
		t.Error("TempDir failed:", err)
		return
	}
	defer os.RemoveAll(pathPrefix)

	err = createFakeDirectoryStructure(pathPrefix, proc)
	if err != nil {
		t.Error("CreateFakeDirectoryStructure failed:", err)
		return
	}

	fakeProcDir := filepath.Join(pathPrefix, "/proc")
	inodes, err := findSocketsOfPid(fakeProcDir, 766)
	if err != nil {
		t.Fatalf("FindSocketsOfPid: %s", err)
	}

	assertUint64ArraysAreEqual(t, []uint64{7619, 7620}, inodes)
}

func TestRefreshCleanupProcessMap(t *testing.T) {
	logp.LogInit(logp.LOG_DEBUG, "", false, true, []string{})

	proc := []testProcFile{
		{path: "/proc/766/fd/10", isLink: true, contents: "/var/log/nginx/packetbeat.error.log"},
		{path: "/proc/766/fd/15", isLink: true, contents: "socket:[7620]"},
	}

	// Create fake proc file system
	pathPrefix, err := ioutil.TempDir("/tmp", "")
	if err != nil {
		t.Error("TempDir failed:", err)
		return
	}
	defer os.RemoveAll(pathPrefix)

	err = createFakeDirectoryStructure(pathPrefix, proc)
	if err != nil {
		t.Error("CreateFakeDirectoryStructure failed:", err)
		return
	}

	// Making new Processes and changing the /proc root to point to fake proc fs
	fakeProcDir := filepath.Join(pathPrefix, "/proc")
	var ps *Processes = New(false, false)
	ps.procFSDirectory = fakeProcDir

	err = ps.Refresh()
	if err != nil {
		t.Fatalf("Refresh failed: %s", err)
	}
	if len(ps.byInode) != 1 {
		t.Errorf("Expected length of ps.byInode is %d but got %d", 1, len(ps.byInode))
	}

	inodes, err := findSocketsOfPid(fakeProcDir, 766)
	if err != nil {
		t.Fatalf("findSocketsOfPid failed: %s", err)
	}
	assertUint64ArraysAreEqual(t, []uint64{7620}, inodes)

	// Removing the process from the file system and refreshing the processes
	processDir := filepath.Join(pathPrefix, "/proc", strconv.Itoa(766))
	os.RemoveAll(processDir)
	err = ps.Refresh()
	if err != nil {
		t.Fatalf("Refresh failed: %s", err)
	}
	if len(ps.byInode) != 0 {
		t.Errorf("Expected length of ps.byInode is %d but got %d", 0, len(ps.byInode))
	}
}

func assertUint64ArraysAreEqual(t *testing.T, expected []uint64, result []uint64) bool {
	for _, ex := range expected {
		found := false
		for _, res := range result {
			if ex == res {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected array %v but got %v", expected, result)
			return false
		}
	}
	return true
}
