package main

import (
	"errors"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func readTree(dir string) (map[string][]string, error) {
	_, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	tree := make(map[string][]string)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		r, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		d := r
		if !info.IsDir() {
			d = filepath.Dir(r)
			r = info.Name()
		}
		if d == r {
			return nil
		}
		tree[d] = append(tree[d], r)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func copyTree(tree map[string][]string, dstHome, srcHome string) error {
	for path, files := range tree {
		for _, f := range files {
			srcDir := filepath.Join(srcHome, path)
			dstDir := filepath.Join(dstHome, path)

			if err := os.MkdirAll(dstDir, 0777); err != nil {
				return err
			}

			src := filepath.Join(srcDir, f)
			dst := filepath.Join(dstDir, f)

			content, err := ioutil.ReadFile(src)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(dst, content, 0777); err != nil {
				return err
			}
		}
	}
	return nil
}

func hash(name string) []byte {
	h := fnv.New64a()
	_, err := h.Write([]byte(name))
	if err != nil {
		panic(err)
	}
	hv := h.Sum(nil)
	for i := 0; i < 4; i++ {
		hv[i] = hv[i] ^ hv[i+4]
	}
	hv = hv[:4]
	return hv
}

func hashParts(hash []byte) []string {
	parts := make([]string, 4)
	for i, v := range hash {
		parts[i] = fmt.Sprintf("%02X", v)
	}
	return parts
}

func validateName(stashName string) bool {
	return regexp.MustCompile("^[a-zA-Z0-9-_]+$").MatchString(stashName)
}

// Errors
var (
	errInvalidStashName = errors.New("invalid stash name")
	errStashNotExist    = errors.New("stash does not exist")
)

func polishStashName(stashName string) string {
	stashName = strings.ToLower(stashName)
	stashName = strings.TrimSpace(stashName)
	return stashName
}

func createStash(stashName, stashTree, fstashHome string) error {
	stashName = polishStashName(stashName)
	if !validateName(stashName) {
		return errInvalidStashName
	}
	tree, err := readTree(stashTree)
	if err != nil {
		return err
	}
	parts := []string{fstashHome}
	parts = append(parts, hashParts(hash(stashName))...)
	parts = append(parts, stashName)
	dst := filepath.Join(parts...)

	return copyTree(tree, dst, stashTree)
}

func expandStash(stashName, fstashHome, workingDirectory string) error {
	stashName = polishStashName(stashName)
	parts := []string{fstashHome}
	parts = append(parts, hashParts(hash(stashName))...)
	parts = append(parts, stashName)
	dir := filepath.Join(parts...)
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return errStashNotExist
		}
		return err
	}

	parts = []string{fstashHome}
	parts = append(parts, hashParts(hash(stashName))...)
	parts = append(parts, stashName)
	stashDir := filepath.Join(parts...)

	tree, err := readTree(stashDir)
	if err != nil {
		return err
	}

	return copyTree(tree, workingDirectory, stashDir)
}

func listDepth(dir string, depth int) ([]string, error) {
	if depth == 0 {
		return nil, nil
	}
	var result []string
	err := filepath.Walk(dir, func(p string, inf os.FileInfo, _err error) error {
		if !inf.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		var parts []string
		for i := 0; i < depth; i++ {
			parts = append(parts, "*")
		}
		matched, err := filepath.Match(filepath.Join(parts...), rel)
		if err != nil {
			return err
		}
		if !matched {
			return nil
		}
		result = append(result, filepath.Base(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
