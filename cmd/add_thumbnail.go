package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/filebrowser/filebrowser/v2/diskcache"
	"github.com/filebrowser/filebrowser/v2/files"
	fbhttp "github.com/filebrowser/filebrowser/v2/http"
)

func init() {
	addThumbnailCmd.Flags().String("cacheDir", "", "file cache directory (required)")
	rootCmd.AddCommand(addThumbnailCmd)
}

var addThumbnailCmd = &cobra.Command{
	Use:   "add-thumbnail <file>",
	Short: "Generate a video thumbnail and store it in the cache",
	Long: `Generates a video thumbnail with the same ffmpeg pipeline the server uses
for video previews and stores it in the disk cache, so the web UI can serve
it without re-encoding.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cacheDir, err := cmd.Flags().GetString("cacheDir")
		if err != nil {
			return err
		}
		return addThumbnail(args[0], cacheDir)
	},
}

var thumbVideoExts = map[string]bool{".mp4": true, ".mkv": true}

func addThumbnail(file, cacheDir string) error {
	if cacheDir == "" {
		return errors.New("--cacheDir is required")
	}

	if !thumbVideoExts[strings.ToLower(filepath.Ext(file))] {
		return errors.New("only .mp4 and .mkv files are supported")
	}

	path, err := filepath.Abs(file)
	if err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}

	out, err := exec.Command("ffmpeg", fbhttp.VideoThumbnailArgs(path)...).Output()
	if err != nil {
		return fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return fmt.Errorf("can't make directory %s: %w", cacheDir, err)
	}

	f := &files.FileInfo{Path: path, ModTime: info.ModTime()}
	cache := diskcache.New(afero.NewOsFs(), cacheDir)
	ctx := context.Background()

	for _, size := range []fbhttp.PreviewSize{fbhttp.PreviewSizeThumb, fbhttp.PreviewSizeBig} {
		key := fbhttp.PreviewCacheKey(f, size)
		if err := cache.Store(ctx, key, out); err != nil {
			return err
		}
		fmt.Printf("thumbnail stored (%s): %s\n", size, cache.Path(key))
	}

	return nil
}
