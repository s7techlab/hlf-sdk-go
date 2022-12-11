package fetcher

import (
	"archive/tar"
	"fmt"
	"io"
	fsPath "path"

	"github.com/go-git/go-billy/v5"
)

func addFileToTar(tw *tar.Writer, path string, fs billy.Filesystem) error {
	files, err := fs.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read dir: %w", err)
	}

	if err = tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     path,
		Mode:     defaultMode,
	}); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	for _, f := range files {
		fPath := fsPath.Join(path, f.Name())
		if f.IsDir() {
			if f.Name()[0:1] == `.` {
				continue
			}
			if err = addFileToTar(tw, fPath, fs); err != nil {
				return fmt.Errorf("failed to write path %s: %w", fPath, err)
			}
			continue
		}

		if err = tw.WriteHeader(&tar.Header{
			Name:    fPath,
			Size:    f.Size(),
			Mode:    defaultMode,
			ModTime: f.ModTime(),
		}); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}
		file, err := fs.Open(fPath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		if _, err = io.Copy(tw, file); err != nil {
			file.Close()
			return fmt.Errorf("failed to write file to tar: %w", err)
		}
		file.Close()
	}
	return nil
}
