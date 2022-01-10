package notedb

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func init() {
	RegisterFormat(FormatDescription{
		ID:     "test",
		Open:   testOpen,
		Detect: testDetect,
	})
}

func testOpen(dbURL *url.URL) (Database, error) {
	return nil, fmt.Errorf("testOpen called with %s", dbURL.String())
}

func testDetect(dbURL *url.URL) DetectResult {
	if strings.HasSuffix(dbURL.Path, ".test") {
		return DetectResultPositive
	}
	return DetectResultNegative
}

func TestOpenDatabase(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		dbURL, err := url.Parse("test:///")
		require.NoError(t, err)
		_, err = OpenDatabase(dbURL)
		require.Error(t, err)
		require.Equal(t, err.Error(), "testOpen called with test:///")
	})
	t.Run("nested scheme", func(t *testing.T) {
		dbURL, err := url.Parse("test+ssh:///")
		require.NoError(t, err)
		_, err = OpenDatabase(dbURL)
		require.Error(t, err)
		require.Equal(t, err.Error(), "testOpen called with test+ssh:///")
	})
}

func TestResolveURL(t *testing.T) {
	t.Run("absolute file path", func(t *testing.T) {
		ret, err := ResolveURL("/a/b")
		require.NoError(t, err)
		require.Equal(t, ret, &url.URL{
			Scheme: "file",
			Path:   "/a/b",
		})
	})
	t.Run("relative file path", func(t *testing.T) {
		abs, err := filepath.Abs("a/b")
		require.NoError(t, err)
		ret, err := ResolveURL("a/b")
		require.NoError(t, err)
		require.Equal(t, ret, &url.URL{
			Scheme: "file",
			Path:   abs,
		})
	})
	t.Run("detected file path", func(t *testing.T) {
		ret, err := ResolveURL("/a.test")
		require.NoError(t, err)
		require.Equal(t, ret, &url.URL{
			Scheme: "test",
			Path:   "/a.test",
		})
	})
	t.Run("URL", func(t *testing.T) {
		ret, err := ResolveURL("test:///a/b")
		require.NoError(t, err)
		require.Equal(t, ret, &url.URL{
			Scheme: "test",
			Path:   "/a/b",
		})
	})
}
