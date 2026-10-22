package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/diskcache"
	"github.com/filebrowser/filebrowser/v2/files"
	fbhttp "github.com/filebrowser/filebrowser/v2/http"
)

func TestAddThumbnailRejectsNonVideo(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(file, []byte("hi"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := addThumbnail(file, filepath.Join(dir, "cache")); err == nil {
		t.Error("expected an error for a non-video file")
	}
}

func TestAddThumbnailStoresBothSizes(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}

	dir := t.TempDir()
	video := filepath.Join(dir, "sample.mp4")
	gen := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i", "testsrc=duration=1:size=128x128:rate=1", video)
	if out, err := gen.CombinedOutput(); err != nil {
		t.Skipf("failed to generate test video: %v: %s", err, out)
	}

	cacheDir := filepath.Join(dir, "cache")
	if err := addThumbnail(video, cacheDir); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(video)
	if err != nil {
		t.Fatal(err)
	}
	f := &files.FileInfo{Path: video, ModTime: info.ModTime()}
	cache := diskcache.New(afero.NewOsFs(), cacheDir)

	for _, size := range []fbhttp.PreviewSize{fbhttp.PreviewSizeThumb, fbhttp.PreviewSizeBig} {
		key := fbhttp.PreviewCacheKey(f, size)
		data, ok, err := cache.Load(t.Context(), key)
		if err != nil {
			t.Fatal(err)
		}
		if !ok || len(data) == 0 {
			t.Errorf("expected cached thumbnail for size %v", size)
		}
		if _, err := os.Stat(cache.Path(key)); err != nil {
			t.Errorf("expected Path(%q) to point to an existing file: %v", key, err)
		}
	}
}
