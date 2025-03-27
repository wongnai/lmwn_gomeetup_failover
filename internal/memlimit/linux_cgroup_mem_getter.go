package memlimit

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrNotSupported = errors.New("not supported")
)

var checkInitSysFSOnce sync.Once

func (l *linuxCgroupMemoryGetter) checkInitSysFS() {
	checkInitSysFSOnce.Do(func() {
		if l.sysFs == nil {
			l.sysFs = os.DirFS("/")
		}
	})
}

type linuxCgroupMemoryGetter struct {
	sysFs fs.FS
}

func (l *linuxCgroupMemoryGetter) GetUsedBytes() (int64, error) {
	l.checkInitSysFS()

	f, err := l.sysFs.Open("sys/fs/cgroup/memory/memory.usage_in_bytes")
	if os.IsNotExist(err) {
		return 0, ErrNotSupported
	} else if err != nil {
		return 0, err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return 0, err
	}
	return parseInt(b)
}

func (l *linuxCgroupMemoryGetter) GetTotalBytes() (int64, error) {
	l.checkInitSysFS()

	f, err := l.sysFs.Open("sys/fs/cgroup/memory/memory.limit_in_bytes")
	if os.IsNotExist(err) {
		return 0, ErrNotSupported
	} else if err != nil {
		return 0, err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return 0, err
	}
	return parseInt(b)
}

func parseInt(b []byte) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
}

func ProvideMemoryGetter() MemoryGetter {
	return &linuxCgroupMemoryGetter{}
}
