package filesystem

import (
	"fmt"
	"sort"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

func TestExcludeFiles(t *testing.T) {
	test0Entry := fmt.Sprintf("%v test0", faker.Sentence())
	test3Entry := fmt.Sprintf("%vtest3%v", faker.Name(), faker.Word())
	fileList := []string{
		test0Entry,
		fmt.Sprintf("test1&£&(%v", faker.URL()),
		fmt.Sprintf("%vtest2%v", faker.URL(), faker.Sentence()),
		test3Entry,
		fmt.Sprintf("%v()T$^test4%v", faker.IPv4(), faker.UUIDHyphenated()),
	}

	tests := []struct {
		files             []string
		exclusionPatterns []string
		expectedError     error
		expectedResult    []string
	}{
		{
			files:             fileList,
			exclusionPatterns: nil,
			expectedError:     nil,
			expectedResult:    fileList,
		},
		{
			files:             fileList,
			exclusionPatterns: []string{".*test0.*", ".*test1.*", ".*test2.*", ".*test3.*", ".*test4.*"},
			expectedError:     nil,
			expectedResult:    []string{},
		},
		{
			files:             fileList,
			exclusionPatterns: []string{".*test1.*", ".*test2.*", ".*test4.*"},
			expectedError:     nil,
			expectedResult:    []string{test0Entry, test3Entry},
		},
		{
			files:             fileList,
			exclusionPatterns: []string{".*test1.*", ".*test2.*", ".*test3.*", ".*test4.*"},
			expectedError:     nil,
			expectedResult:    []string{test0Entry},
		},
		{
			files:             fileList,
			exclusionPatterns: []string{".*test0.*", ".*test1.*", ".*test2.*", ".*test4.*"},
			expectedError:     nil,
			expectedResult:    []string{test3Entry},
		},
		{
			files:             fileList,
			exclusionPatterns: []string{"*test0**"},
			expectedError:     commonerrors.ErrInvalid,
			expectedResult:    []string{},
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("exlusions [%v]", test.exclusionPatterns), func(t *testing.T) {
			regexes, err := NewExclusionRegexList(globalFileSystem.PathSeparator(), test.exclusionPatterns...)
			if test.expectedError != nil {
				require.Error(t, err)
				assert.True(t, commonerrors.Any(err, test.expectedError))
			} else {
				newList, err := ExcludeFiles(test.files, regexes)
				require.NoError(t, err)
				sort.Strings(newList)
				sort.Strings(test.expectedResult)
				assert.Equal(t, test.expectedResult, newList)
				for i := range newList {
					assert.False(t, IsPathExcludedFromPatterns(newList[i], globalFileSystem.PathSeparator(), test.exclusionPatterns...))
				}
			}

		})
	}
}

func TestExcludes(t *testing.T) {
	t.Parallel() // marks TLog as capable of running in parallel with other tests
	var listOfPaths = []string{
		"some/path", "somepath", ".snapshot", ".snapshot/path", "test/.snapshot/some/path", ".snapshot123", ".snapshot123/path", "test/.snapshot123/some-path", "test/.snapshot123/some/path",
	}
	tests := []struct {
		inputlist       []string
		exclusions      []string
		expectedResults []string
	}{
		{
			inputlist:       listOfPaths,
			exclusions:      []string{},
			expectedResults: listOfPaths,
		},
		{
			inputlist:       listOfPaths,
			exclusions:      []string{"noexclusion"},
			expectedResults: listOfPaths,
		},
		{
			inputlist:       []string{},
			exclusions:      []string{"any"},
			expectedResults: []string{},
		},
		{
			inputlist:       listOfPaths,
			exclusions:      []string{""},
			expectedResults: listOfPaths,
		},
		{
			inputlist:       listOfPaths,
			exclusions:      []string{"some.*"},
			expectedResults: []string{".snapshot", ".snapshot/path", ".snapshot123", ".snapshot123/path"},
		},
		{
			inputlist:       listOfPaths,
			exclusions:      []string{".*path"},
			expectedResults: []string{".snapshot", ".snapshot123"},
		},
		{
			inputlist:       listOfPaths,
			exclusions:      []string{"[.]snapshot.*"},
			expectedResults: []string{"some/path", "somepath"},
		},
		{
			inputlist:       listOfPaths,
			exclusions:      []string{"[.]snapshot.*/.*"},
			expectedResults: []string{"some/path", "somepath", ".snapshot", ".snapshot123"},
		},
		{
			inputlist:       listOfPaths,
			exclusions:      []string{"[.]snapshot.*", ".*path"},
			expectedResults: []string{},
		},
	}
	for i := range tests {
		test := tests[i] // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(fmt.Sprintf("%v: %v", i, test.exclusions), func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			actualList, err := ExcludeAll(test.inputlist, test.exclusions...)
			require.NoError(t, err)
			assert.Equal(t, test.expectedResults, actualList)
		})
	}

}

func TestIsPathExcludedFromPatterns(t *testing.T) {
	tests := []struct {
		path              string
		pathSeparator     rune
		exclusionPatterns []string
		exclude           bool
	}{
		{
			path:              "",
			pathSeparator:     '\\',
			exclusionPatterns: []string{},
			exclude:           false,
		},
		{
			path:              "",
			pathSeparator:     '\\',
			exclusionPatterns: []string{".*[.]test[^13]"},
			exclude:           false,
		},
		{
			path:              "C:\\Users\\adrcab01\\AppData\\Local\\Temp\\test-findall-929837903\\level1\\level2\\test-findall-309750873.test2",
			pathSeparator:     '\\',
			exclusionPatterns: []string{},
			exclude:           false,
		},
		{
			path:              "C:\\Users\\adrcab01\\AppData\\Local\\Temp\\test-findall-929837903\\level1\\level2\\test-findall-309750873.test2",
			pathSeparator:     '\\',
			exclusionPatterns: []string{".*[.]test[^13]"},
			exclude:           true,
		},
		{
			path:              "C:/Users/adrcab01/AppData/Local/Temp/test-findall-929837903/level1/level2/test-findall-309750873.test2",
			pathSeparator:     '/',
			exclusionPatterns: []string{".*[.]test[^13]"},
			exclude:           true,
		},
		{
			path:              "C:\\Users\\adrcab01\\AppData\\Local\\Temp\\test-findall-929837903\\level1\\level2\\test-findall-309750873.test3",
			pathSeparator:     '\\',
			exclusionPatterns: []string{".*[.]test[^13]"},
			exclude:           false,
		},
		{
			path:              "C:/Users/adrcab01/AppData/Local/Temp/test-findall-929837903/level1/level2/test-findall-309750873.test3",
			pathSeparator:     '/',
			exclusionPatterns: []string{".*[.]test[^13]"},
			exclude:           false,
		},
		{
			path:              "C:\\Users\\adrcab01\\AppData\\Local\\Temp\\test-findall-929837903\\level1\\level2\\test-findall-309750873.test2",
			pathSeparator:     '\\',
			exclusionPatterns: []string{".*[.]test2"},
			exclude:           true,
		},
		{
			path:              "C:\\Users\\adrcab01\\AppData\\Local\\Temp\\test-findall-929837903\\level1\\level2\\test-findall-309750873.test3",
			pathSeparator:     '\\',
			exclusionPatterns: []string{".*[.]test2"},
			exclude:           false,
		},
		{
			path:              "C:\\Users\\adrcab01\\AppData\\Local\\Temp\\test-findall-929837903\\level1\\level2\\test-findall-309750873.test3",
			pathSeparator:     '\\',
			exclusionPatterns: []string{".*"},
			exclude:           true,
		},
		{
			path:              "C:\\Users\\adrcab01\\AppData\\Local\\Temp\\test-findall-929837903\\level1\\level2\\test-findall-309750873.test3",
			pathSeparator:     '\\',
			exclusionPatterns: []string{".*test3"},
			exclude:           true,
		},
		{
			path:              "C:\\Users\\adrcab01\\AppData\\Local\\Temp\\test-findall-929837903\\level1\\level2\\test-findall-309750873.test3",
			pathSeparator:     '\\',
			exclusionPatterns: []string{".*test2"},
			exclude:           false,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("exlusions [%v]", test.exclusionPatterns), func(t *testing.T) {
			if test.exclude {
				assert.True(t, IsPathExcludedFromPatterns(test.path, test.pathSeparator, test.exclusionPatterns...))
			} else {
				assert.False(t, IsPathExcludedFromPatterns(test.path, test.pathSeparator, test.exclusionPatterns...))
			}

		})
	}
}
