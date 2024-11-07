package filesystem

import (
	"archive/zip"
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"go.uber.org/atomic"
	"golang.org/x/text/encoding/unicode"

	"github.com/ARM-software/golang-utils/utils/charset"
	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/parallelisation"
	"github.com/ARM-software/golang-utils/utils/safecast"
	"github.com/ARM-software/golang-utils/utils/safeio"
)

const (
	zipExt                = ".zip"
	sevenzipExt           = ".7z"
	sevenzipmacExt        = ".s7z"
	gzipExt               = ".gz"
	lzipExt               = ".lz"
	zipxExt               = ".zipx"
	targzExt              = ".tar.gz"
	targz2Ext             = ".tgz"
	xzExt                 = ".xz"
	lzmaExt               = ".lzma"
	rzipExt               = ".rz"
	packExt               = ".pack"
	compressExt           = ".z"
	jarExt                = ".jar"
	zipMimeType           = "application/zip"
	zipxMimeType          = "application/x-zip"
	zipCompressedMimeType = "application/x-zip-compressed"
	jarMimeType           = "application/jar"
	epubMimeType          = "application/epub+zip"
)

var (
	// ZipFileExtensions returns a list of commonly used extensions to describe zip archive files
	// This list was populated from [Wikipedia](https://en.wikipedia.org/wiki/List_of_archive_formats)
	ZipFileExtensions = []string{zipExt, zipxExt, sevenzipExt, sevenzipmacExt, gzipExt, targzExt, targz2Ext, xzExt, lzipExt, lzmaExt, rzipExt, packExt, compressExt, jarExt}
	// ZipMimeTypes returns a list of MIME types describing zip archive files.
	ZipMimeTypes = []string{zipMimeType, zipxMimeType, zipCompressedMimeType, jarMimeType, epubMimeType}
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
		n, err := safeio.CopyDataWithContext(ctx, src, dest)
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

// Prevents any ZipSlip ([CWE-22](https://cwe.mitre.org/data/definitions/22.html)) (files outside extraction dirPath) https://snyk.io/research/zip-slip-vulnerability#go
func sanitiseZipExtractPath(fs FS, filePath string, destination string) (destPath string, err error) {
	destPath = filepath.Join(destination, filePath) // join cleans the destpath so we can check for ZipSlip
	if destPath == destination {
		return
	}
	if !strings.Contains(destPath, "..") {
		if strings.HasPrefix(destPath, fmt.Sprintf("%v%v", destination, string(fs.PathSeparator()))) {
			return
		}
		if strings.HasPrefix(destPath, fmt.Sprintf("%v/", destination)) {
			return
		}
	}

	err = fmt.Errorf("%w: zipslip security breach detected, file dirPath '%s' not in destination directory '%s'", commonerrors.ErrMalicious, filePath, destination)
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
	fileList, _, _, err = fs.unzip(ctx, source, destination, NoLimits(), 0)
	return
}

// UnzipWithContextAndLimits unzips an source archive file into destination.
func UnzipWithContextAndLimits(ctx context.Context, source string, destination string, limits ILimits) ([]string, error) {
	return globalFileSystem.UnzipWithContextAndLimits(ctx, source, destination, limits)
}

func (fs *VFS) UnzipWithContextAndLimits(ctx context.Context, source string, destination string, limits ILimits) (fileList []string, err error) {
	fileList, _, _, err = fs.unzip(ctx, source, destination, limits, 0)
	return
}

func newZipReader(fs FS, source string, limits ILimits, currentDepth int64) (zipReader *zip.Reader, file File, err error) {
	if fs == nil {
		err = fmt.Errorf("%w: missing file system", commonerrors.ErrUndefined)
		return
	}
	if limits == nil {
		err = fmt.Errorf("%w: missing file system limits", commonerrors.ErrUndefined)
		return
	}
	if limits.Apply() && limits.GetMaxDepth() >= 0 && currentDepth > limits.GetMaxDepth() {
		err = fmt.Errorf("%w: depth [%v] of zip file [%v] is beyond allowed limits (max: %v)", commonerrors.ErrTooLarge, currentDepth, source, limits.GetMaxDepth())
		return
	}

	if !fs.Exists(source) {
		err = fmt.Errorf("%w: could not find archive [%v]", commonerrors.ErrNotFound, source)
		return
	}

	info, err := fs.Lstat(source)
	if err != nil {
		return
	}
	file, err = fs.GenericOpen(source)
	if err != nil {
		return
	}

	zipFileSize := info.Size()

	if limits.Apply() && zipFileSize > limits.GetMaxFileSize() {
		err = fmt.Errorf("%w: zip file [%v] is too big (%v B) and beyond limits (max: %v B)", commonerrors.ErrTooLarge, source, zipFileSize, limits.GetMaxFileSize())
		return
	}

	zipReader, err = zip.NewReader(file, zipFileSize)
	err = convertZipError(err)

	return
}

func (fs *VFS) unzip(ctx context.Context, source string, destination string, limits ILimits, currentDepth int64) (fileList []string, fileOnDiskCount uint64, sizeOnDisk uint64, err error) {

	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	fileCounter := atomic.NewUint64(0)

	// List of file paths to return
	totalSizeOnDisk := atomic.NewUint64(0)

	zipReader, f, err := newZipReader(fs, source, limits, currentDepth)
	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()
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
		subErr := parallelisation.DetermineContextError(ctx)
		if subErr != nil {
			return fileList, fileCounter.Load(), totalSizeOnDisk.Load(), subErr
		}
		zippedFile := zipReader.File[i]
		// Detection of Zip slip https://cwe.mitre.org/data/definitions/22.html (CodeQL)
		if strings.Contains(zippedFile.Name, "..") {
			_, subErr := sanitiseZipExtractPath(fs, zippedFile.Name, destination)
			if subErr != nil {
				return fileList, fileCounter.Load(), totalSizeOnDisk.Load(), subErr
			}
		}
		// Calculate file dirPath
		filePath, subErr := sanitiseZipExtractPath(fs, zippedFile.Name, destination)
		if subErr != nil {
			return fileList, fileCounter.Load(), totalSizeOnDisk.Load(), subErr
		}

		var fileDepth int64
		if limits.Apply() && limits.GetMaxDepth() >= 0 {
			depth, subErr := FileTreeDepth(fs, destination, filePath)
			fileDepth = depth + currentDepth
			if subErr != nil {
				return fileList, fileCounter.Load(), totalSizeOnDisk.Load(), subErr
			}
			if fileDepth > limits.GetMaxDepth() {
				subErr = fmt.Errorf("%w: depth [%v] of file [%v] within zip [%v] is beyond allowed limits (max: %v)", commonerrors.ErrTooLarge, fileDepth, filepath.Base(filePath), filepath.Base(source), limits.GetMaxDepth())
				return fileList, fileCounter.Load(), totalSizeOnDisk.Load(), subErr
			}
		}

		// record unzipped files (except zip files if they get unzipped later)
		if !(limits.ApplyRecursively() && fs.isZipWithContext(ctx, zippedFile.Name)) {
			fileCounter.Inc()
			fileList = append(fileList, filePath)
		}

		if zippedFile.FileInfo().IsDir() {
			// Create directory
			subErr = fs.MkDir(filePath)
			if subErr != nil {
				return fileList, fileCounter.Load(), totalSizeOnDisk.Load(), fmt.Errorf("unable to create directory [%s]: %w", filePath, subErr)
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
			return fileList, fileCounter.Load(), totalSizeOnDisk.Load(), fmt.Errorf("unable to create directory '%s': %w", directoryPath, subErr)
		}

		fileSizeOnDisk, subErr := fs.unzipZippedFile(ctx, filePath, zippedFile, limits, fileDepth)
		if subErr != nil {
			return fileList, fileCounter.Load(), totalSizeOnDisk.Load(), subErr
		}

		// If the copied file is a zip, unzip that zip if the action is marked as recursive
		if limits.ApplyRecursively() {
			if fs.isZipWithContext(ctx, filePath) {
				nestedUnzippedFiles, filesOnDiskCount, filesSizeOnDisk, subErr := fs.unzipNestedZipFiles(ctx, filePath, limits, fileDepth)
				if subErr != nil {
					return fileList, fileCounter.Load(), totalSizeOnDisk.Load(), subErr
				}
				totalSizeOnDisk.Add(filesSizeOnDisk)
				fileCounter.Add(filesOnDiskCount)
				fileList = append(fileList, nestedUnzippedFiles...)
			} else {
				if fs.isZipWithContext(ctx, zippedFile.Name) { // If not an actual zip file but with a zip name.
					fileCounter.Inc()
					fileList = append(fileList, filePath)
				}
				totalSizeOnDisk.Add(safecast.ToUint64(fileSizeOnDisk))
			}
		} else {
			totalSizeOnDisk.Add(safecast.ToUint64(fileSizeOnDisk))
		}

		if limits.Apply() && totalSizeOnDisk.Load() > limits.GetMaxTotalSize() {
			return fileList, fileCounter.Load(), totalSizeOnDisk.Load(), fmt.Errorf("%w: more than %v B of disk space was used while unzipping %v (%v B used already)", commonerrors.ErrTooLarge, limits.GetMaxTotalSize(), source, totalSizeOnDisk.Load())
		}
		if filecount := fileCounter.Load(); limits.Apply() && filecount <= math.MaxInt64 && safecast.ToInt64(filecount) > limits.GetMaxFileCount() {
			return fileList, filecount, totalSizeOnDisk.Load(), fmt.Errorf("%w: more than %v files were created while unzipping %v (%v files created already)", commonerrors.ErrTooLarge, limits.GetMaxFileCount(), source, filecount)
		}
	}

	// Ensuring directory timestamps are preserved (this needs to be done after all the files have been created).
	err = preserveDirectoriesTimestamps(ctx, fs, directoryInfo)
	if err != nil {
		return fileList, fileCounter.Load(), totalSizeOnDisk.Load(), err
	}

	return fileList, fileCounter.Load(), totalSizeOnDisk.Load(), nil
}

func (fs *VFS) unzipNestedZipFiles(ctx context.Context, nestedZipFile string, limits ILimits, currentDepth int64) (nestedUnzippedFiles []string, fileOnDiskCount uint64, filesSizeOnDisk uint64, err error) {
	destination := filepath.Join(filepath.Dir(nestedZipFile), FilepathStem(nestedZipFile))
	nestedUnzippedFiles, fileOnDiskCount, filesSizeOnDisk, subErr := fs.unzip(ctx, nestedZipFile, destination, limits, currentDepth+1)
	if subErr != nil {
		err = fmt.Errorf("unable to unzip nested zip [%s] present at depth (%d) to [%s] : %w", filepath.Base(nestedZipFile), currentDepth, destination, subErr)
		return
	}
	subErr = fs.Rm(nestedZipFile)
	if subErr != nil {
		err = fmt.Errorf("unable to remove nested zip [%s] : %w", nestedZipFile, subErr)
	}
	return
}

func preserveDirectoriesTimestamps(ctx context.Context, fs FS, directoryInfo map[string]os.FileInfo) error {
	for dirPath, dirInfo := range directoryInfo {
		subErr := parallelisation.DetermineContextError(ctx)
		if subErr != nil {
			return subErr
		}
		times := newDefaultTimeInfo(dirInfo)
		subErr = fs.Chtimes(dirPath, times.AccessTime(), times.ModTime())
		if subErr != nil {
			return fmt.Errorf("unable to set directory timestamp [%s]: %w", dirPath, subErr)
		}
	}
	return nil
}

// unzipZippedFile unzips file to destination directory
func (fs *VFS) unzipZippedFile(ctx context.Context, dest string, zippedFile *zip.File, limits ILimits, currentDepth int64) (fileSizeOnDisk int64, err error) {
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}

	if limits.Apply() && limits.GetMaxDepth() > 0 && currentDepth > limits.GetMaxDepth() {
		err = fmt.Errorf("%w: depth [%v] of zipped file [%v] is beyond allowed limits (max: %v)", commonerrors.ErrTooLarge, currentDepth, zippedFile.Name, limits.GetMaxDepth())
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

	_, err = safeio.CopyNWithContext(ctx, sourceFile, destinationFile, fileSizeOnDisk)
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

func (fs *VFS) IsZip(path string) bool {
	return fs.isZipWithContext(context.Background(), path)
}

func (fs *VFS) isZipWithContext(ctx context.Context, path string) bool {
	found, _ := fs.IsZipWithContext(ctx, path)
	return found
}

func (fs *VFS) IsZipWithContext(ctx context.Context, path string) (ok bool, err error) {
	if path == "" {
		err = fmt.Errorf("%w: missing path", commonerrors.ErrUndefined)
		return
	}
	err = parallelisation.DetermineContextError(ctx)
	if err != nil {
		return
	}
	_, found := collection.Find(&ZipFileExtensions, strings.ToLower(filepath.Ext(path)))
	if !found || err != nil {
		return
	}
	if !fs.Exists(path) {
		ok = found
		return
	}
	f, err := fs.GenericOpen(path)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	content, err := safeio.ReadAll(ctx, f)
	if err != nil {
		return
	}
	mime := http.DetectContentType(content)
	_, ok = collection.Find(&ZipMimeTypes, mime)
	return
}

func convertZipError(err error) error {
	if err == nil {
		return nil
	}
	err = commonerrors.ConvertContextError(err)
	if commonerrors.Any(err, zip.ErrFormat, zip.ErrChecksum) {
		return fmt.Errorf("%w: %v", commonerrors.ErrInvalid, err.Error())
	}
	if commonerrors.Any(err, zip.ErrFormat, zip.ErrAlgorithm) {
		return fmt.Errorf("%w: %v", commonerrors.ErrUnsupported, err.Error())
	}
	return err
}
