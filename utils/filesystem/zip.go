package filesystem

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"go.uber.org/atomic"
	"golang.org/x/text/encoding/unicode"

	"github.com/ARM-software/golang-utils/utils/charset"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
)

const (
	zipExt = ".zip"
)

// Zip zips a source directory into a destination archive
func Zip(source string, destination string) error {
	return globalFileSystem.Zip(source, destination)
}

func (fs *VFS) Zip(source, destination string) error {
	return fs.ZipWithContext(context.Background(), source, destination)
}

func (fs *VFS) ZipWithContext(ctx context.Context, source, destination string) (err error) {
	return fs.ZipWithContextAndLimits(ctx, source, destination, NoLimits())
}

func (fs *VFS) ZipWithContextAndLimits(ctx context.Context, source, destination string, limits ILimits) error {
	return fs.ZipWithContextAndLimitsAndExclusionPatterns(ctx, source, destination, limits)
}

func (fs *VFS) ZipWithContextAndLimitsAndExclusionPatterns(ctx context.Context, source string, destination string, limits ILimits, exclusionPatterns ...string) (err error) {
	if limits == nil {
		err = fmt.Errorf("%w: missing file system limits", commonerrors.ErrUndefined)
		return
	}

	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	file, err := fs.CreateFile(destination)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	// create a new zip archive
	w := zip.NewWriter(file)
	defer func() { _ = w.Close() }()

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if limits.Apply() && info.Size() > limits.GetMaxFileSize() {
			err = fmt.Errorf("%w: file [%v] is too big (%v B) and beyond limits (max: %v B)", commonerrors.ErrTooLarge, path, info.Size(), limits.GetMaxFileSize())
			return err
		}

		// Get the relative path
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		// If directory
		if info.IsDir() {
			if path == source {
				return nil
			}
			header := &zip.FileHeader{
				Name:     relPath + "/",
				Method:   zip.Deflate,
				Modified: info.ModTime(),
			}
			_, err = w.CreateHeader(header)
			return err
		}

		// if file
		src, err := fs.GenericOpen(path)
		if err != nil {
			return err
		}
		defer func() { _ = src.Close() }()

		// create file in archive
		relPath, err = filepath.Rel(source, path)
		if err != nil {
			return err
		}
		header := &zip.FileHeader{
			Name:     relPath,
			Method:   zip.Deflate,
			Modified: info.ModTime(),
		}
		dest, err := w.CreateHeader(header)
		if err != nil {
			return err
		}
		n, err := io.Copy(dest, src)
		if err != nil {
			return err
		}

		if info.Size() != n && !IsSymLink(info) {
			return fmt.Errorf("could not write the full file [%v] content (wrote %v/%v bytes)", relPath, n, info.Size())
		}
		return nil
	}
	err = fs.WalkWithContextAndExclusionPatterns(ctx, source, walker, exclusionPatterns...)

	if limits.Apply() {
		stat, subErr := file.Stat()
		if subErr != nil {
			return subErr
		}
		if stat.Size() > limits.GetMaxFileSize() {
			subErr = fmt.Errorf("%w: file [%v] is too big (%v B) and beyond limits (max: %v B)", commonerrors.ErrTooLarge, destination, stat.Size(), limits.GetMaxFileSize())
			return subErr
		}
	}
	return
}

// Prevents any ZipSlip (files outside extraction dirPath) https://snyk.io/research/zip-slip-vulnerability#go
func sanitiseZipExtractPath(fs FS, filePath string, destination string) (destPath string, err error) {
	destPath = filepath.Join(destination, filePath) // join cleans the destpath so we can check for ZipSlip
	if destPath == destination {
		return
	}
	if strings.HasPrefix(destPath, fmt.Sprintf("%v%v", destination, string(fs.PathSeparator()))) {
		return
	}
	if strings.HasPrefix(destPath, fmt.Sprintf("%v/", destination)) {
		return
	}
	err = fmt.Errorf("%w: zipslip security breach detected, file dirPath '%s' not in destination directory '%s'", commonerrors.ErrInvalidDestination, filePath, destination)
	return
}

// Unzip unzips an source archive file into destination.
func Unzip(source, destination string) ([]string, error) {
	return globalFileSystem.Unzip(source, destination)
}

func (fs *VFS) Unzip(source, destination string) ([]string, error) {
	return fs.UnzipWithContext(context.Background(), source, destination)
}

func (fs *VFS) UnzipWithContext(ctx context.Context, source string, destination string) (fileList []string, err error) {
	return fs.unzip(ctx, source, destination, NoLimits(), 0)
}

func (fs *VFS) UnzipWithContextAndLimits(ctx context.Context, source string, destination string, limits ILimits) (fileList []string, err error) {
	return fs.unzip(ctx, source, destination, limits, 0)
}

func (fs *VFS) unzip(ctx context.Context, source string, destination string, limits ILimits, depth int64) (fileList []string, err error) {
	if limits == nil {
		err = fmt.Errorf("%w: missing file system limits", commonerrors.ErrUndefined)
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	if limits.Apply() && depth > limits.GetMaxZipDepth() {
		err = fmt.Errorf("%w: zip file [%v] is contains too many nested zipped directories (%v) and beyond limits (max: %v)", commonerrors.ErrTooLarge, source, depth, limits.GetMaxZipDepth())
		return
	}

	fileCounter := atomic.NewUint64(0)

	// List of file paths to return
	totalSizeOnDisk := atomic.NewUint64(0)

	info, err := fs.Lstat(source)
	if err != nil {
		return
	}
	f, err := fs.GenericOpen(source)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	zipFileSize := info.Size()

	if limits.Apply() && zipFileSize > limits.GetMaxFileSize() {
		err = fmt.Errorf("%w: zip file [%v] is too big (%v B) and beyond limits (max: %v B)", commonerrors.ErrTooLarge, source, zipFileSize, limits.GetMaxFileSize())
		return
	}

	zipReader, err := zip.NewReader(f, zipFileSize)
	if err != nil {
		return
	}

	// Clean the destination to find shortest dirPath
	destination = filepath.Clean(destination)
	err = fs.MkDir(destination)
	if err != nil {
		return
	}
	directoryInfo := map[string]os.FileInfo{}

	// For each file in the zip file
	for i := range zipReader.File {
		fileCounter.Inc()

		zippedFile := zipReader.File[i]
		subErr := parallelisation.DetermineContextError(ctx)
		if subErr != nil {
			return fileList, subErr
		}

		// Calculate file dirPath
		filePath, subErr := sanitiseZipExtractPath(fs, zippedFile.Name, destination)
		if subErr != nil {
			return fileList, subErr
		}

		// Keep list of files unzipped (except zip files as they will be handled later)
		if filepath.Ext(zippedFile.Name) != zipExt {
			fileList = append(fileList, filePath)
		}

		if zippedFile.FileInfo().IsDir() {
			// Create directory
			subErr = fs.MkDir(filePath)

			if subErr != nil {
				return fileList, fmt.Errorf("unable to create directory [%s]: %w", filePath, subErr)
			}
			// recording directory dirInfo to preserve timestamps
			directoryInfo[filePath] = zippedFile.FileInfo()
			// Nothing more to do for a directory, move to next zip file
			continue
		}

		// If a file create the dirPath into which to write the file
		directoryPath := filepath.Dir(filePath)
		subErr = fs.MkDir(directoryPath)
		if subErr != nil {
			return fileList, fmt.Errorf("unable to create directory '%s': %w", directoryPath, subErr)
		}

		fileSizeOnDisk, subErr := fs.unzipZipFile(ctx, filePath, zippedFile, limits)
		if subErr != nil {
			return fileList, subErr
		}

		// If file that was copied is a zip, unzip that zip
		if filepath.Ext(zippedFile.Name) == zipExt {
			defer func() { _ = fs.Rm(filePath) }()

			nestedUnzipFiles, subErr := fs.unzip(ctx, filePath, strings.TrimSuffix(filePath, zipExt), limits, depth+1)
			if subErr != nil {
				return fileList, fmt.Errorf("unable to unzip nested zip [%s] to [%s] at depth (%d): %w", filePath, strings.TrimSuffix(filePath, zipExt), depth, subErr)
			}

			fileList = append(fileList, nestedUnzipFiles...)
			continue
		}

		totalSizeOnDisk.Add(uint64(fileSizeOnDisk))
		if limits.Apply() && totalSizeOnDisk.Load() > limits.GetMaxTotalSize() {
			return fileList, fmt.Errorf("%w: more than %v B of disk space was used while unzipping %v (%v B used already)", commonerrors.ErrTooLarge, limits.GetMaxTotalSize(), source, totalSizeOnDisk.Load())
		}
		if limits.Apply() && int64(fileCounter.Load()) > limits.GetMaxFileCount() {
			return fileList, fmt.Errorf("%w: more than %v files were created while unzipping %v (%v files created already)", commonerrors.ErrTooLarge, limits.GetMaxFileCount(), source, fileCounter)
		}
	}

	// Ensuring directory timestamps are preserved (this needs to be done after all the files have been created).
	for dirPath, dirInfo := range directoryInfo {
		subErr := parallelisation.DetermineContextError(ctx)
		if subErr != nil {
			return fileList, subErr
		}
		times := newDefaultTimeInfo(dirInfo)
		subErr = fs.Chtimes(dirPath, times.AccessTime(), times.ModTime())
		if subErr != nil {
			return fileList, fmt.Errorf("unable to set directory timestamp [%s]: %w", dirPath, subErr)
		}
	}

	return fileList, nil
}

// unzipZipFile unzips file to destination directory
func (fs *VFS) unzipZipFile(ctx context.Context, dest string, zippedFile *zip.File, limits ILimits) (fileSizeOnDisk int64, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	destinationPath, err := determineUnzippedFilepath(dest)
	if err != nil {
		return
	}

	destinationFile, err := fs.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zippedFile.Mode())
	if err != nil {
		err = fmt.Errorf("%w: unable to open file '%s': %v", commonerrors.ErrUnexpected, destinationPath, err.Error())
		return
	}
	defer func() { _ = destinationFile.Close() }()

	sourceFile, err := zippedFile.Open()
	if err != nil {
		err = fmt.Errorf("%w: unable to open zipped file '%s': %v", commonerrors.ErrUnsupported, zippedFile.Name, err.Error())
		return
	}
	defer func() { _ = sourceFile.Close() }()

	info := zippedFile.FileInfo()
	fileSizeOnDisk = info.Size()
	if limits.Apply() {
		if fileSizeOnDisk > limits.GetMaxFileSize() {
			err = fmt.Errorf("%w: zipped file [%v] is too big (%v B) and above max size (%v B)", commonerrors.ErrTooLarge, info.Name(), info.Size(), limits.GetMaxFileSize())
			return
		}
	}

	_, err = io.CopyN(destinationFile, sourceFile, fileSizeOnDisk)
	if err != nil {
		err = fmt.Errorf("copy of zipped file to '%s' failed: %w", destinationPath, err)
		return
	}
	err = destinationFile.Close()
	if err != nil {
		return
	}
	// Ensuring the timestamp is preserved.
	times := newDefaultTimeInfo(info)
	err = fs.Chtimes(destinationPath, times.AccessTime(), times.ModTime())
	// Nothing more to do for a directory, move to next zip file
	return
}

func determineUnzippedFilepath(destinationPath string) (string, error) {

	// See https://go-review.googlesource.com/c/go/+/75592/
	// Character encodings other than CP-437 and UTF-8
	// are not officially supported by the ZIP specification, pragmatically
	// the world has permitted use of them.
	//
	// When a non-standard encoding is used, it is the user's responsibility
	// to ensure that the target system is expecting the encoding used
	// (e.g., producing a ZIP file you know is used on a Chinese version of Windows).
	if utf8.ValidString(destinationPath) {
		return destinationPath, nil
	}
	// Nonetheless, instead of raising an error when non-UTF8 characters are present in filepath,
	// we try to guess the encoding and then, convert the string to UTF-8.
	encoding, charsetName, err := charset.DetectTextEncoding([]byte(destinationPath))
	if err != nil {
		return "", fmt.Errorf("%w: file path [%s] is not a valid utf-8 string and charset could not be detected: %v", commonerrors.ErrInvalid, destinationPath, err.Error())
	}
	convertedDestinationPath, err := charset.IconvString(destinationPath, encoding, unicode.UTF8)
	if err != nil {
		return "", fmt.Errorf("%w: file path [%s] is encoded using charset [%v] but could not be converted to valid utf-8: %v", commonerrors.ErrUnexpected, destinationPath, charsetName, err.Error())
		// If zip file paths must be accepted even when their encoding is unknown, or conversion to utf-8 failed, then the following can be done.
		// destinationPath = strings.ToValidUTF8(dest, charset.InvalidUTF8CharacterReplacement)
	}
	return convertedDestinationPath, err
}
