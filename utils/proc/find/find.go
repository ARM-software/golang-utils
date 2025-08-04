package find

import (
	"context"
	"fmt"
	"regexp"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/proc"
)

const numWorkers = 10

// FindProcessByRegex will search for the processes that match a specific regex
func FindProcessByRegex(ctx context.Context, re *regexp.Regexp) (processes []proc.IProcess, err error) {
	if re == nil {
		err = commonerrors.UndefinedVariable("regex to search")
		return
	}
	return findProcessByRegex(ctx, re)
}

// FindProcessByName will search for the processes that match a specific name
func FindProcessByName(ctx context.Context, name string) (processes []proc.IProcess, err error) {
	return FindProcessByRegex(ctx, regexp.MustCompile(fmt.Sprintf(".*%v.*", regexp.QuoteMeta(name))))
}
