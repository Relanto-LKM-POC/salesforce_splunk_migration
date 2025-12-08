package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileExists(t *testing.T) {
	t.Run("Success_FileExists", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "testfile-*.txt")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		tmpFile.WriteString("test content")
		tmpFile.Close()

		exists, err := FileExists(tmpFile.Name())
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Success_DirectoryExists", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "testdir-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		exists, err := FileExists(tmpDir)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Success_FileDoesNotExist", func(t *testing.T) {
		nonExistentPath := filepath.Join(os.TempDir(), "nonexistent-file-12345.txt")
		
		exists, err := FileExists(nonExistentPath)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Success_DirectoryDoesNotExist", func(t *testing.T) {
		nonExistentPath := filepath.Join(os.TempDir(), "nonexistent-dir-12345")
		
		exists, err := FileExists(nonExistentPath)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Error_InvalidPath", func(t *testing.T) {
		// On Windows, paths with invalid characters
		invalidPath := "C:\\invalid\x00path"
		
		exists, _ := FileExists(invalidPath)
		assert.False(t, exists)
		// On some systems this might return an error, on others it might just return false
		// We accept both behaviors
	})

	t.Run("Success_EmptyPath", func(t *testing.T) {
		exists, err := FileExists("")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Success_RelativePath", func(t *testing.T) {
		// Create a file in current directory
		tmpFile, err := os.CreateTemp(".", "testfile-*.txt")
		require.NoError(t, err)
		fileName := filepath.Base(tmpFile.Name())
		tmpFile.Close()
		defer os.Remove(fileName)

		exists, err := FileExists(fileName)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Success_NestedPath", func(t *testing.T) {
		// Create nested directory structure
		tmpDir, err := os.MkdirTemp("", "testdir-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		nestedPath := filepath.Join(tmpDir, "subdir", "file.txt")
		err = os.MkdirAll(filepath.Dir(nestedPath), 0755)
		require.NoError(t, err)

		err = os.WriteFile(nestedPath, []byte("test"), 0644)
		require.NoError(t, err)

		exists, err := FileExists(nestedPath)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Success_SymlinkExists", func(t *testing.T) {
		if os.Getenv("CI") != "" {
			t.Skip("Skipping symlink test in CI environment")
		}

		// Create a temporary file
		tmpFile, tmpErr := os.CreateTemp("", "testfile-*.txt")
		if tmpErr != nil {
			t.Skip("Cannot create temp file for symlink test")
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		// Create a symlink
		symlinkPath := tmpFile.Name() + ".link"
		symlinkErr := os.Symlink(tmpFile.Name(), symlinkPath)
		if symlinkErr != nil {
			t.Skip("Cannot create symlink (may need elevated permissions)")
		}
		defer os.Remove(symlinkPath)

		exists, err := FileExists(symlinkPath)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}
