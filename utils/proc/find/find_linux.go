//go:build linux

package find

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/proc"
	"github.com/ARM-software/golang-utils/utils/safecast"
)

const (
	procFS        = "/proc"
	procDataFile  = "cmdline"
	invalidPIDErr = "could not parse PID from proc path"
)

func checkProcessMatch(ctx context.Context, fs filesystem.FS, re *regexp.Regexp, procEntry string) (ok bool, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	data, err := fs.ReadFile(procEntry)
	if err != nil {
		if commonerrors.CorrespondTo(err, "no bytes were read", "no such file or directory") {
			err = nil
			return // ignore special descriptors since our cmdline will have content (we still have to check since all files in proc have size zero) as well as processes that have stopped since the listing occurred
		}
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not read proc entry '%v'", procEntry)
		return
	}

	data = bytes.ReplaceAll(data, []byte{0}, []byte{' '}) // https://man7.org/linux/man-pages/man5/proc_pid_cmdline.5.html

	ok = re.Match(data)
	return
}

func parseProcess(ctx context.Context, entry string) (p proc.IProcess, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	pidStr := strings.Trim(strings.TrimSuffix(strings.TrimPrefix(entry, procFS), fmt.Sprintf("%v", procDataFile)), "/")
	// proc.FindProcess eventually converts the PID with safecast.ToInt32 before calling gopsutil.
	// An out-of-range PID would be clamped: one above MaxInt32 would be looked up
	// as MaxInt32, potentially returning the wrong process or an incorrect not-found
	// result. Parse as int32 here to reject the invalid PID before any lookup.
	// gopsutil accepts int32 PIDs: https://github.com/shirou/gopsutil/blob/81236e0193191e972dfdf1e15cc1ae53764dc1e0/process/process_posix.go#L103.
	// Linux documents pid_t as signed (https://man7.org/linux/man-pages/man3/pid_t.3type.html)
	// and defines __kernel_pid_t as int (https://github.com/torvalds/linux/blob/master/include/uapi/asm-generic/posix_types.h).
	// See also https://github.com/ARM-software/golang-utils/security/code-scanning/126 and https://pkg.go.dev/strconv#ParseInt.
	pid32, err := strconv.ParseInt(pidStr, 10, 32)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "%v '%v': PID '%v' must be a signed 32-bit decimal integer", invalidPIDErr, entry, pidStr)
		return
	}

	pid := safecast.ToInt(pid32)
	p, err = proc.FindProcess(ctx, pid)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "could not find process '%v'", pid)
		return
	}

	return
}

// FindProcessByRegexForFS will search a given filesystem for the processes that match a specific regex
func FindProcessByRegexForFS(ctx context.Context, fs filesystem.FS, re *regexp.Regexp) (processes []proc.IProcess, err error) {
	if !filesystem.Exists(procFS) {
		err = commonerrors.Newf(commonerrors.ErrNotFound, "the proc filesystem was not found at '%v'", procFS)
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	searchGlobTerm := fmt.Sprintf("%v/[0-9]*/%v", procFS, procDataFile)
	procEntries, err := fs.Glob(searchGlobTerm)
	if err != nil {
		err = commonerrors.WrapErrorf(commonerrors.ErrUnexpected, err, "an error occurred when searching for processes using the following glob '%v'", searchGlobTerm)
		return
	}

	processes, err = parallelisation.WorkerPool(ctx, numWorkers, procEntries, func(ctx context.Context, entry string) (p proc.IProcess, matches bool, err error) {
		matches, err = checkProcessMatch(ctx, fs, re, entry)
		if err != nil || !matches {
			return
		}

		p, err = parseProcess(ctx, entry)
		if err != nil {
			err = commonerrors.IgnoreCorrespondTo(err, invalidPIDErr)
			return
		}

		matches = true
		return
	})

	return
}

// findProcessByRegex will search for the processes that match a specific regex
func findProcessByRegex(ctx context.Context, re *regexp.Regexp) (processes []proc.IProcess, err error) {
	return FindProcessByRegexForFS(ctx, filesystem.GetGlobalFileSystem(), re)
}
