package filesystem

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ARM-software/golang-utils/utils/reflection"
)

// FilepathStem returns  the final path component, without its suffix.
func FilepathStem(fp string) string {
	return strings.TrimSuffix(filepath.Base(fp), filepath.Ext(fp))
}

// FileTreeDepth returns the depth of a file in a tree starting from root
func FileTreeDepth(fs FS, root, filePath string) (depth int64, err error) {
	if reflection.IsEmpty(filePath) {
		return
	}
	rel, err := fs.ConvertToRelativePath(root, filePath)
	if err != nil {
		return
	}
	diff := rel[0]
	if reflection.IsEmpty(diff) {
		return
	}
	fmt.Println(diff)
	diff = strings.ReplaceAll(diff, string(fs.PathSeparator()), "/")
	depth = int64(len(strings.Split(diff, "/")) - 1)
	return
}
