package semver

import (
	"context"
	"embed"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/filesystem"
)

var (
	//go:embed testdata/*
	embeddedFS embed.FS
)

func TestDetermineVersion(t *testing.T) {
	t.Run("from file", func(t *testing.T) {
		libraryVersion := filepath.Join("..", DefaultVersionFile)
		version, err := DetermineVersionFromFile(context.Background(), filesystem.GetGlobalFileSystem(), libraryVersion)
		require.NoError(t, err)
		assert.NotEmpty(t, version)
	})

	t.Run("embedded", func(t *testing.T) {
		version, err := DetermineEmbeddedVersion(context.Background(), &embeddedFS, field.ToOptionalString(fmt.Sprintf("testdata/%v", DefaultVersionFile)))
		require.NoError(t, err)
		assert.NotEmpty(t, version)
		assert.Equal(t, "1.0.1-test", version)
	})
	t.Run("invalid", func(t *testing.T) {
		version, err := DetermineEmbeddedVersion(context.Background(), &embeddedFS, field.ToOptionalString(fmt.Sprintf("testdata/%v", "invalid.properties")))
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		assert.Empty(t, version)
		version, err = DetermineVersionFromFile(context.Background(), filesystem.GetGlobalFileSystem(), filepath.Join("testdata", "invalid2.properties"))
		errortest.AssertError(t, err, commonerrors.ErrInvalid, commonerrors.ErrConflict)
		assert.Empty(t, version)
	})
	t.Run("not found", func(t *testing.T) {
		version, err := DetermineEmbeddedVersion(context.Background(), &embeddedFS, nil)
		errortest.AssertError(t, err, commonerrors.ErrNotFound)
		assert.Empty(t, version)
		version, err = DetermineVersionFromFile(context.Background(), filesystem.GetGlobalFileSystem(), "       ")
		errortest.AssertError(t, err, commonerrors.ErrNotFound, commonerrors.ErrUndefined)
		assert.Empty(t, version)
		version, err = DetermineVersionFromFile(context.Background(), filesystem.GetGlobalFileSystem(), faker.UUIDDigit())
		errortest.AssertError(t, err, commonerrors.ErrNotFound)
		assert.Empty(t, version)
	})
}
