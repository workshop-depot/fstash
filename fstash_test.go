package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if err := createSampleTree(); err != nil {
		panic(err)
	}
	exitCode := m.Run()

	dirs := []string{homeDir1(), homeDir2(), homeDir3(), homeDir4()}
	for _, v := range dirs {
		if err := os.RemoveAll(v); err != nil {
			panic(err)
		}
	}

	os.Exit(exitCode)
}

func createSampleTree() error {
	home := homeDir1()
	d0f1 := filepath.Join(home, "file1.txt")
	d0f2 := filepath.Join(home, "file2.txt")

	dir1 := filepath.Join(home, "dir1")
	d1f1 := filepath.Join(dir1, "file1.txt")
	d1f2 := filepath.Join(dir1, "file2.txt")

	dir2 := filepath.Join(home, "dir2")
	d2f2 := filepath.Join(dir2, "file2.txt")
	d2f1 := filepath.Join(dir2, "file1.txt")

	dir3 := filepath.Join(dir2, "dir3")
	d3f2 := filepath.Join(dir3, "file2.txt")
	d3f1 := filepath.Join(dir3, "file1.txt")

	// create

	if err := os.MkdirAll(dir1, 0777); err != nil {
		return err
	}
	if err := os.MkdirAll(dir2, 0777); err != nil {
		return err
	}
	if err := os.MkdirAll(dir3, 0777); err != nil {
		return err
	}

	content := []byte(strings.Repeat("a", 100))
	files := []string{d0f1, d0f2, d1f1, d1f2, d2f1, d2f2, d3f1, d3f2}
	for _, v := range files {
		if err := ioutil.WriteFile(v, content, 0777); err != nil {
			return err
		}
	}

	return nil
}

func homeDir1() string {
	tmp := os.TempDir()
	return filepath.Join(tmp, "fstash-test-dir-1")
}

func homeDir2() string {
	tmp := os.TempDir()
	return filepath.Join(tmp, "fstash-test-dir-2")
}

func homeDir3() string {
	tmp := os.TempDir()
	return filepath.Join(tmp, "fstash-test-dir-3")
}

func homeDir4() string {
	tmp := os.TempDir()
	return filepath.Join(tmp, "fstash-test-dir-4")
}

func makeOutput(tree map[string][]string) (*strings.Builder, error) {
	sb := new(strings.Builder)
	var keys []string
	for k := range tree {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		_, err := sb.WriteString(fmt.Sprintf("%v %v\n", k, tree[k]))
		if err != nil {
			return nil, err
		}
	}
	return sb, nil
}

func Test_createSampleTree(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(createSampleTree())
}

func Test_readTree(t *testing.T) {
	assert := assert.New(t)

	tree, err := readTree(homeDir1())
	assert.NoError(err)

	sb, err := makeOutput(tree)
	assert.NoError(err)

	assert.Equal(`. [file1.txt file2.txt]
dir1 [file1.txt file2.txt]
dir2 [file1.txt file2.txt]
dir2/dir3 [file1.txt file2.txt]
`, sb.String())
}

func Test_copyTree(t *testing.T) {
	assert := assert.New(t)

	tree, err := readTree(homeDir1())
	assert.NoError(err)

	err = copyTree(tree, homeDir2(), homeDir1())
	assert.NoError(err)

	tree, err = readTree(homeDir2())
	assert.NoError(err)

	sb, err := makeOutput(tree)
	assert.NoError(err)

	assert.Equal(`. [file1.txt file2.txt]
dir1 [file1.txt file2.txt]
dir2 [file1.txt file2.txt]
dir2/dir3 [file1.txt file2.txt]
`, sb.String())
}

func Test_copyTree_check_content(t *testing.T) {
	assert := assert.New(t)

	tree, err := readTree(homeDir1())
	assert.NoError(err)

	err = copyTree(tree, homeDir2(), homeDir1())
	assert.NoError(err)

	tree, err = readTree(homeDir2())
	assert.NoError(err)

	content := strings.Repeat("a", 100)
	for path, files := range tree {
		for _, f := range files {
			dir := filepath.Join(homeDir2(), path)
			fp := filepath.Join(dir, f)
			c, err := ioutil.ReadFile(fp)
			assert.NoError(err)
			assert.Equal(content, string(c))
		}
	}
}

func Test_hash(t *testing.T) {
	require := require.New(t)

	h := hash("stash-name")
	require.Len(h, 4)
	require.Equal("DBB22753", fmt.Sprintf("%X", h))
}

func Test_hash_parts(t *testing.T) {
	require := require.New(t)

	h := hash("stash-name")
	require.Len(h, 4)
	require.Equal("DBB22753", fmt.Sprintf("%X", h))

	parts := hashParts(h)
	require.Len(parts, 4)
	require.Equal("[DB B2 27 53]", fmt.Sprint(parts))
}

func Test_validate_stash_name(t *testing.T) {
	require := require.New(t)

	t.Run("some invalid characters", func(t *testing.T) {
		chars := []string{"/", ".", "~", "+", "'"}
		for _, v := range chars {
			require.False(validateName(v))
		}
	})

	t.Run("some valid names", func(t *testing.T) {
		require.True(validateName("some-name"))
		require.True(validateName("some_name"))
		require.True(validateName("1some_name"))
	})

	t.Run("some invalid names", func(t *testing.T) {
		require.False(validateName("some:name"))
		require.False(validateName("(some:name"))
		require.False(validateName("someپname"))
	})
}

func Test_stash_directory_create_new_stash_validate_stash_name(t *testing.T) {
	require := require.New(t)

	stashTree := homeDir1()
	fstashHome := homeDir3()
	stashName := "sample-stash" + "::"
	err := createStash(stashName, stashTree, fstashHome)
	require.Equal(errInvalidStashName, err)
}

func Test_stash_directory_create_new_stash(t *testing.T) {
	require := require.New(t)

	stashTree := homeDir1()
	fstashHome := homeDir3()
	stashName := "sample-stash"
	err := createStash(stashName, stashTree, fstashHome)
	require.NoError(err)

	parts := []string{fstashHome}
	parts = append(parts, hashParts(hash(stashName))...)
	parts = append(parts, stashName)
	dir := filepath.Join(parts...)
	tree, err := readTree(dir)
	require.NoError(err)

	sb, err := makeOutput(tree)
	require.NoError(err)

	require.Equal(`. [file1.txt file2.txt]
dir1 [file1.txt file2.txt]
dir2 [file1.txt file2.txt]
dir2/dir3 [file1.txt file2.txt]
`, sb.String())
}

func Test_pop_stash_not_exist(t *testing.T) {
	require := require.New(t)

	stashName := "sample-stash"
	fstashHome := homeDir1()
	workingDirectory := homeDir4()
	err := popStash(stashName, fstashHome, workingDirectory)
	require.Equal(errStashNotExist, err)
}

func Test_pop_stash(t *testing.T) {
	require := require.New(t)

	stashName := "sample-stash"
	fstashHome := homeDir3()
	workingDirectory := homeDir4()

	err := popStash(stashName, fstashHome, workingDirectory)
	require.NoError(err)

	tree, err := readTree(workingDirectory)
	require.NoError(err)

	sb, err := makeOutput(tree)
	require.NoError(err)

	require.Equal(`. [file1.txt file2.txt]
dir1 [file1.txt file2.txt]
dir2 [file1.txt file2.txt]
dir2/dir3 [file1.txt file2.txt]
`, sb.String())
}

func Test_listDepth(t *testing.T) {
	require := require.New(t)

	stashTree := homeDir1()
	fstashHome := homeDir3()

	{
		stashName := "sample-stash-1"
		err := createStash(stashName, stashTree, fstashHome)
		require.NoError(err)
	}

	{
		stashName := "sample-stash-2"
		err := createStash(stashName, stashTree, fstashHome)
		require.NoError(err)
	}

	{
		stashName := "sample-stash-3"
		err := createStash(stashName, stashTree, fstashHome)
		require.NoError(err)
	}

	l, err := listDepth(fstashHome, 5)
	require.NoError(err)
	require.Len(l, 4)
	require.Equal("[sample-stash sample-stash-1 sample-stash-2 sample-stash-3]",
		fmt.Sprint(l))
}
