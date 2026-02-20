package platform

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveWithPrivileges(t *testing.T) {
	t.Skip("Can only be run as administrator")
	tmpDir := t.TempDir()
	defer func() { _ = os.RemoveAll(tmpDir) }()

	dir, err := os.MkdirTemp(tmpDir, "testDirToRemove")
	require.NoError(t, err)

	assert.NoError(t, RemoveWithPrivileges(context.TODO(), dir))

	_, err = os.Stat(dir)
	assert.Error(t, err)

	f, err := os.CreateTemp(tmpDir, "testfile")
	require.NoError(t, err)
	err = f.Close()
	require.NoError(t, err)

	assert.NoError(t, RemoveWithPrivileges(context.TODO(), f.Name()))

	_, err = os.Stat(f.Name()) //nolint:gosec //G703
	assert.Error(t, err)
}
