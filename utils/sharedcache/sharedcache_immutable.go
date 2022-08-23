package sharedcache

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/filesystem"
	"github.com/ARM-software/golang-utils/utils/idgen"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

const (
	partFileDescriptor     = ".part"
	defaultCachedPackageID = "cb93fdbe-6c2e-4f7d-96ac-c422fc52618e "
)

type SharedImmutableCacheRepository struct {
	AbstractSharedCacheRepository
}

type FileWithModTime struct {
	filename string
	modTime  time.Time
}

func NewSharedImmutableCacheRepository(cfg *Configuration, fs filesystem.FS) (repository *SharedImmutableCacheRepository, err error) {
	abstractCache, err := NewAbstractSharedCacheRepository(cfg, fs)
	if err != nil {
		return
	}
	repository = &SharedImmutableCacheRepository{
		AbstractSharedCacheRepository: *abstractCache,
	}
	return
}

func listCompleteFilesByModTime(ctx context.Context, fs filesystem.FS, entryDir string) (sorted []string, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	// sort the files by mod time
	var fileModTimes []FileWithModTime
	files, err := fs.Ls(entryDir)
	if err != nil {
		return
	}

	// Create array of non .part and non .hash files with their modtimes
	for _, file := range files {
		err = parallelisation.DetermineContextError(ctx)
		if err != nil {
			return sorted, err
		}
		isPartFile := strings.EqualFold(filepath.Ext(file), partFileDescriptor)
		isHashFile := strings.EqualFold(filepath.Ext(file), hashFileDescriptor)
		if !isPartFile && !isHashFile {
			fullPath := filepath.Join(entryDir, file)
			statInfo, err := fs.StatTimes(fullPath)
			if err != nil {
				return sorted, err
			}
			fileModTime := statInfo.ModTime()
			fileModTimes = append(fileModTimes, FileWithModTime{
				filename: file,
				modTime:  fileModTime,
			})
		}
	}
	// can't do this directly using strings and stattimes to check on the
	// file because the sort.Slice requires the function to output bool
	// so if we wouldn't be abele to return an error if stattimes failed
	sort.Slice(fileModTimes, func(i, j int) bool { return fileModTimes[i].modTime.After(fileModTimes[j].modTime) })

	// map to just string
	for _, file := range fileModTimes {
		err = parallelisation.DetermineContextError(ctx)
		if err != nil {
			return sorted, err
		}
		sorted = append(sorted, file.filename)
	}

	return
}

func (s *SharedImmutableCacheRepository) Fetch(ctx context.Context, key, dest string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	// setup the lock
	remoteDir := s.getCacheEntryPath(key)
	remoteExists, err := s.fs.IsDir(remoteDir)
	if err != nil {
		return
	}
	if !remoteExists {
		err = fmt.Errorf("no cache entry for key [%v]: %w", key, commonerrors.ErrNotFound)
		return
	}
	err = s.setUpLocalDestination(ctx, dest)
	if err != nil {
		return err
	}
	// find the most recent cached package
	cachedPackage, err := s.findCachedPackageFromEntryDir(ctx, key, remoteDir)
	if err != nil {
		return err
	}
	err = s.unpackPackageToLocalDestination(ctx, cachedPackage, dest)
	return
}

func (s *SharedImmutableCacheRepository) Store(ctx context.Context, key, src string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	remoteDir, err := s.createEntry(ctx, key)
	if err != nil {
		return
	}

	// create temp location for files so we don't include zip inside itself
	tempDir, err := s.fs.TempDirInTempDir(tempDirPrefix)
	if err != nil {
		return
	}
	defer func() { _ = s.fs.Rm(tempDir) }()

	// zip the local cache
	// generate zip
	zipped := filepath.Join(tempDir, s.generateCachedPackageName())
	err = s.fs.ZipWithContext(ctx, src, zipped)
	if err != nil {
		return
	}

	// do the transfer
	destZip, err := TransferFiles(ctx, s.fs, remoteDir, zipped)
	if err != nil {
		_ = s.fs.Rm(destZip)
		return
	}

	// remove .part from uploaded cache file
	if strings.EqualFold(filepath.Ext(destZip), partFileDescriptor) {
		finalZip := strings.ReplaceAll(destZip, partFileDescriptor, "")
		err = s.fs.Move(destZip, finalZip)
		// Don't forget the hash file
		hashFile := filepath.Join(filepath.Dir(destZip), generateHashFileName(destZip))
		if s.fs.Exists(hashFile) {
			finalHash := strings.ReplaceAll(hashFile, partFileDescriptor, "")
			_ = s.fs.Move(hashFile, finalHash)
		}
	}
	return
}

func (s *SharedImmutableCacheRepository) generateCachedPackageName() string {
	cacheUUID, err := idgen.GenerateUUID4()
	if err != nil {
		cacheUUID = defaultCachedPackageID
	}
	return fmt.Sprintf("%v-%v%v", cacheUUID, defaultCachedPackage, partFileDescriptor)
}

func (s *SharedImmutableCacheRepository) CleanEntry(ctx context.Context, key string) (err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	entryDir := s.getCacheEntryPath(key)
	// need to remove all files in cache except most recent
	files, err := listCompleteFilesByModTime(ctx, s.fs, entryDir)
	if err != nil {
		return
	}
	if len(files) < 2 {
		return
	}
	toClean := files[1:]
	for _, file := range toClean {
		err = parallelisation.DetermineContextError(ctx)
		if err != nil {
			return err
		}
		packageFile := filepath.Join(entryDir, file)
		err = s.fs.Rm(packageFile)
		if err != nil {
			return err
		}
		// Don't forget the hash file
		hashFile := filepath.Join(filepath.Dir(packageFile), generateHashFileName(packageFile))
		if s.fs.Exists(hashFile) {
			_ = s.fs.Rm(hashFile)
		}
	}
	// clean cache ignore .part files as it might be run at the same time as an upload
	return
}

func (s *SharedImmutableCacheRepository) findCachedPackageFromEntryDir(ctx context.Context, key, entryDir string) (cachedPackage string, err error) {
	// find the most recent cached package
	files, err := listCompleteFilesByModTime(ctx, s.fs, entryDir)
	if err != nil {
		return
	}
	if len(files) == 0 {
		err = fmt.Errorf("no entry for key [%v] in cache: %w", key, commonerrors.ErrEmpty)
		return
	}
	cachedPackage = filepath.Join(entryDir, files[0])
	return
}

func (s *SharedImmutableCacheRepository) GetEntryAge(ctx context.Context, key string) (age time.Duration, err error) {
	return s.getEntryAge(ctx, key, s.findCachedPackageFromEntryDir)
}

func (s *SharedImmutableCacheRepository) SetEntryAge(ctx context.Context, key string, age time.Duration) error {
	return s.setEntryAge(ctx, key, age, s.findCachedPackageFromEntryDir)
}
