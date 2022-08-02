package util

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CreateTar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer
func CreateTar(src, output string) error {
	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf(fmt.Sprintf("Unable to tar files - %v", err))
	}

	outputPath := output
	if !strings.HasSuffix(outputPath, ".tar.gz") {
		outputPath = outputPath + ".tar.gz"
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error writing archive: %v", err)
	}
	defer out.Close()

	gzw := gzip.NewWriter(out)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// walk path
	return filepath.WalkDir(src, func(file string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		dInfo, err := d.Info()
		if err != nil {
			return err
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(dInfo, d.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()

		return nil
	})
}
