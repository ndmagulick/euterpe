package fileio

import (
	"fmt"
	"os"
	"path/filepath"
)

func PathsOverlap(path1, path2 string) (bool, error) {
	if path1 == "" || path2 == "" {
		return false, fmt.Errorf("pathsOverlap: empty path entered")
	}

	resolvedPath1, err := resolvePath(path1)
	if err != nil {
		return false, fmt.Errorf("resolving %q: %w", path1, err)
	}

	resolvedPath2, err := resolvePath(path2)
	if err != nil {
		return false, fmt.Errorf("resolving %q: %w", path2, err)
	}

	return pathIsWithin(resolvedPath1, resolvedPath2) || pathIsWithin(resolvedPath2, resolvedPath1), nil
}

func ValidateDirectoryPath(path string) error {
	pathInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("validateDirectoryPath %q: %w", path, err)
	}

	if !pathInfo.IsDir() {
		return fmt.Errorf("validateDirectoryPath: path is not a directory: %q", path)
	}

	return nil
}

func GetExecutableDirectory() (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("getting executable path: %w", err)
	}

	resolvedPath, err := filepath.EvalSymlinks(executablePath)
	if err != nil {
		return "", fmt.Errorf("evaluating sym links for executable path: %w", err)
	}

	return filepath.Dir(resolvedPath), nil
}

func resolvePath(path string) (string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return filepath.EvalSymlinks(absolutePath)
}

func pathIsWithin(parentPath, childPath string) bool {
	relativePath, err := filepath.Rel(parentPath, childPath)
	if err != nil {
		return false
	}

	return filepath.IsLocal(relativePath)
}
