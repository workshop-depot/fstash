package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	os.Exit(m.Run())
}

func createSampleTree(home string) error {
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

func randTemp() string {
	const (
		testDirPrefix = "fstash-test-dir-"
	)

	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), testDirPrefix+hex.EncodeToString(randBytes))
}

func Test_createSampleTree(t *testing.T) {
	require := require.New(t)
	homeDir := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir))
	}()

	require.Nil(createSampleTree(homeDir))
}

func Test_readTree(t *testing.T) {
	require := require.New(t)
	homeDir := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir))
	}()

	require.Nil(createSampleTree(homeDir))

	tree, err := readTree(homeDir)
	require.NoError(err)

	sb, err := makeOutput(tree)
	require.NoError(err)

	require.Equal(`. [file1.txt file2.txt]
dir1 [file1.txt file2.txt]
dir2 [file1.txt file2.txt]
dir2/dir3 [file1.txt file2.txt]
`, sb.String())
}

func Test_copyTree(t *testing.T) {
	require := require.New(t)
	homeDir1 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir1))
	}()
	homeDir2 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir2))
	}()

	require.Nil(createSampleTree(homeDir1))

	tree, err := readTree(homeDir1)
	require.NoError(err)

	err = copyTree(tree, homeDir2, homeDir1)
	require.NoError(err)

	tree, err = readTree(homeDir2)
	require.NoError(err)

	sb, err := makeOutput(tree)
	require.NoError(err)

	require.Equal(`. [file1.txt file2.txt]
dir1 [file1.txt file2.txt]
dir2 [file1.txt file2.txt]
dir2/dir3 [file1.txt file2.txt]
`, sb.String())
}

func Test_copyTree_check_content(t *testing.T) {
	require := require.New(t)
	homeDir1 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir1))
	}()
	homeDir2 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir2))
	}()

	require.Nil(createSampleTree(homeDir1))

	tree, err := readTree(homeDir1)
	require.NoError(err)

	err = copyTree(tree, homeDir2, homeDir1)
	require.NoError(err)

	tree, err = readTree(homeDir2)
	require.NoError(err)

	content := strings.Repeat("a", 100)
	for path, files := range tree {
		for _, f := range files {
			dir := filepath.Join(homeDir2, path)
			fp := filepath.Join(dir, f)
			c, err := ioutil.ReadFile(fp)
			require.NoError(err)
			require.Equal(content, string(c))
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
	homeDir1 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir1))
	}()
	homeDir3 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir3))
	}()

	require.Nil(createSampleTree(homeDir1))

	stashTree := homeDir1
	fstashHome := homeDir3
	stashName := "sample-stash" + "::"
	err := createStash(stashName, stashTree, fstashHome)
	require.Equal(errInvalidStashName, err)
}

func Test_stash_directory_create_new_stash(t *testing.T) {
	require := require.New(t)
	homeDir1 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir1))
	}()
	homeDir3 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir3))
	}()

	require.Nil(createSampleTree(homeDir1))

	stashTree := homeDir1
	fstashHome := homeDir3
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

func Test_expand_stash_not_exist(t *testing.T) {
	require := require.New(t)
	homeDir1 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir1))
	}()
	homeDir4 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir4))
	}()

	stashName := "sample-stash"
	fstashHome := homeDir1
	workingDirectory := homeDir4
	err := expandStash(stashName, fstashHome, workingDirectory, nil)
	require.Equal(errStashNotExist, err)
}

func Test_expand_stash(t *testing.T) {
	require := require.New(t)
	homeDir3 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir3))
	}()
	homeDir4 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir4))
	}()
	homeDir1 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir1))
	}()

	require.Nil(createSampleTree(homeDir1))

	// creating sample stash
	{
		stashTree := homeDir1
		fstashHome := homeDir3
		stashName := "sample-stash"
		err := createStash(stashName, stashTree, fstashHome)
		require.NoError(err)
	}

	stashName := "sample-stash"
	fstashHome := homeDir3
	workingDirectory := homeDir4

	err := expandStash(stashName, fstashHome, workingDirectory, nil)
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
	homeDir3 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir3))
	}()
	homeDir1 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir1))
	}()

	require.Nil(createSampleTree(homeDir1))

	stashTree := homeDir1
	fstashHome := homeDir3

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
	require.Len(l, 3)
	require.Equal("[sample-stash-1 sample-stash-2 sample-stash-3]",
		fmt.Sprint(l))
}

const (
	staticContent   = "some static content"
	templateContent = "Author of {{ .AppName }} is {{ .Author }}."
)

func createSampleTreeWithTemplates(home string) error {
	if err := os.MkdirAll(home, 0777); err != nil {
		return err
	}

	d0f1 := filepath.Join(home, "file1.txt")
	d0f2 := filepath.Join(home, "file2.txt")

	dir1 := filepath.Join(home, "dir1")
	d1f3 := filepath.Join(dir1, "file3.txt")
	d1f4 := filepath.Join(dir1, "file4.txt")

	if err := os.MkdirAll(dir1, 0777); err != nil {
		return err
	}

	if err := ioutil.WriteFile(d0f1, []byte(staticContent), 0777); err != nil {
		return err
	}
	if err := ioutil.WriteFile(d0f2, []byte(templateContent), 0777); err != nil {
		return err
	}

	if err := ioutil.WriteFile(d1f3, []byte(staticContent), 0777); err != nil {
		return err
	}
	if err := ioutil.WriteFile(d1f4, []byte(templateContent), 0777); err != nil {
		return err
	}

	return nil
}

func Test_expand_stash_templates(t *testing.T) {
	require := require.New(t)
	homeDir3 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir3))
	}()
	homeDir4 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir4))
	}()
	homeDir1 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir1))
	}()

	// creating sample stash
	{
		require.Nil(createSampleTreeWithTemplates(homeDir1))

		stashTree := homeDir1
		fstashHome := homeDir3
		stashName := "sample-stash"
		err := createStash(stashName, stashTree, fstashHome)
		require.NoError(err)
	}

	// expand stash
	{
		stashName := "sample-stash"
		fstashHome := homeDir3
		workingDirectory := homeDir4

		data := map[string]string{
			"file2": `{"AppName":"fstash","Author":"dc0d"}`,
			"file4": `{"AppName":"Web","Author":"Web Developer"}`,
		}
		err := expandStash(stashName, fstashHome, workingDirectory, data)
		require.NoError(err)
	}

	// check content
	{
		workingDirectory := homeDir4
		d0f1 := filepath.Join(workingDirectory, "file1.txt")
		d0f2 := filepath.Join(workingDirectory, "file2.txt")

		dir1 := filepath.Join(workingDirectory, "dir1")
		d1f3 := filepath.Join(dir1, "file3.txt")
		d1f4 := filepath.Join(dir1, "file4.txt")

		content, err := ioutil.ReadFile(d0f1)
		require.NoError(err)
		require.Equal(staticContent, string(content))

		content, err = ioutil.ReadFile(d0f2)
		require.NoError(err)
		require.Equal("Author of fstash is dc0d.", string(content))

		content, err = ioutil.ReadFile(d1f3)
		require.NoError(err)
		require.Equal(staticContent, string(content))

		content, err = ioutil.ReadFile(d1f4)
		require.NoError(err)
		require.Equal("Author of Web is Web Developer.", string(content))
	}
}

func Test_deleteStash(t *testing.T) {
	require := require.New(t)
	homeDir3 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir3))
	}()
	homeDir1 := filepath.Join(os.TempDir(), randTemp())
	defer func() {
		require.NoError(os.RemoveAll(homeDir1))
	}()

	// creating sample stash
	{
		require.Nil(createSampleTreeWithTemplates(homeDir1))

		stashTree := homeDir1
		fstashHome := homeDir3
		stashName := "sample-stash"
		err := createStash(stashName, stashTree, fstashHome)
		require.NoError(err)
	}

	stashName := "sample-stash"
	fstashHome := homeDir3
	if err := deleteStash(stashName, fstashHome); err != nil {
		require.NoError(err)
	}

	l, err := listDepth(fstashHome, 5)
	require.NoError(err)
	require.Len(l, 0)
	require.Equal("[]", fmt.Sprint(l))
}
