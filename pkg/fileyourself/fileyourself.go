package fileyourself

import (
	"io/fs"
	"os"
	stdFp "path/filepath"
	"strings"
)

const (
	Separator     = os.PathSeparator
	ListSeparator = os.PathListSeparator
)

//nolint:unused // This is a wrapper around the stdlib's filepath package.
var (
	ErrBadPattern = stdFp.ErrBadPattern

	skipAll = fs.SkipAll
	SkipDir = fs.SkipDir
)

func Abs(path string) (string, error) {
	return stdFp.Abs(path)
}

func Base(path string) string {
	return stdFp.Base(path)
}

func Clean(path string) string {
	return stdFp.Clean(path)
}

func Dir(path string) string {
	return stdFp.Dir(path)
}

func EvalSymlinks(path string) (string, error) {
	return stdFp.EvalSymlinks(path)
}

func Ext(path string) string {
	e := stdFp.Ext(path)
	if path == e {
		return ""
	}

	return strings.TrimPrefix(e, ".")
}

func FromSlash(path string) string {
	return stdFp.FromSlash(path)
}

func Glob(pattern string) ([]string, error) {
	return stdFp.Glob(pattern)
}

// HasPrefix is deprecated and is not included in this package.

func IsAbs(path string) bool {
	return stdFp.IsAbs(path)
}

func IsLocal(path string) bool {
	return stdFp.IsLocal(path)
}

func Join(elem ...string) string {
	return stdFp.Join(elem...)
}

func Match(pattern, name string) (bool, error) {
	return stdFp.Match(pattern, name)
}

func Rel(basepath, targpath string) (string, error) {
	return stdFp.Rel(basepath, targpath)
}

func Split(path string) (string, string) {
	return stdFp.Split(path)
}

func SplitList(path string) []string {
	return stdFp.SplitList(path)
}

func ToSlash(path string) string {
	return stdFp.ToSlash(path)
}

func VolumeName(path string) string {
	return stdFp.VolumeName(path)
}

func Walk(root string, fn stdFp.WalkFunc) error {
	return stdFp.Walk(root, fn)
}

func WalkDir(root string, fn fs.WalkDirFunc) error {
	return stdFp.WalkDir(root, fn)
}

type WalkFunc func(path string, info fs.FileInfo, err error) error
