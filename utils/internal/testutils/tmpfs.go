package testutils

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/idgen"
)

func CreateTestFileTree(fs filesystem.FS, testDir string, fileModTime time.Time, fileAccessTime time.Time) ([]string, error) {
	// This can be fixed for testing.
	random := rand.New(rand.NewSource(time.Now().Unix())) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec as this is just for
	err := fs.MkDir(testDir)
	if err != nil {
		return nil, err
	}
	randI := random.Intn(5) //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
	if randI == 0 {
		randI = 1
	}
	for i := 0; i < randI; i++ {
		c := fmt.Sprintf("test%v", i+1)
		path := filepath.Join(testDir, c)

		err = fs.MkDir(path)
		if err != nil {
			return nil, err
		}
		for j := 0; j < random.Intn(5); j++ { //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
			uuid, err := idgen.GenerateUUID4()
			if err != nil {
				uuid = "uuid"
			}
			c := fmt.Sprintf("test-%v-%v", uuid, j+1)
			path := filepath.Join(path, c)

			err = fs.MkDir(path)
			if err != nil {
				return nil, err
			}
			for k := 0; k < random.Intn(5); k++ { //nolint:gosec //causes G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec), So disable gosec
				uuid, err = idgen.GenerateUUID4()
				if err != nil {
					uuid = "uuid"
				}
				c := fmt.Sprintf("test-%v-%v%v", uuid, k+1, ".txt")
				finalPath := filepath.Join(path, c)

				s := fmt.Sprintf("file-%v-%v%v%v ", uuid, i+1, j+1, k+1)
				err = fs.WriteFile(finalPath, []byte(s), 0755)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	var tree []string
	err = fs.ListDirTree(testDir, &tree)
	if err != nil {
		return nil, err
	}

	// unifying timestamps
	for _, path := range tree {
		err = fs.Chtimes(path, fileAccessTime, fileModTime)
		if err != nil {
			return nil, err
		}
	}

	return tree, nil
}
