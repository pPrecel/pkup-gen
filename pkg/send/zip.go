package send

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
)

type zipper struct {
	logger *pterm.Logger
}

func NewZipper(logger *pterm.Logger) *zipper {
	return &zipper{
		logger: logger,
	}
}

func (z *zipper) Do(zipFilename, zipPrefix string) (string, error) {
	zipFilenameBase := filepath.Base(zipFilename)
	zipFilenameDir := filepath.Dir(zipFilename)
	fullZipPath := fmt.Sprintf("%s/%s%s.%s", zipFilenameDir, zipPrefix, zipFilenameBase, "zip")

	z.logger.Debug("creating zip file", z.logger.Args(
		"path", fullZipPath,
	))
	outFile, err := os.Create(fullZipPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	w := zip.NewWriter(outFile)
	defer w.Close()

	baseInZip := fmt.Sprintf("%s%s", zipPrefix, zipFilenameBase)
	if err := z.addFilesToZip(w, zipFilename, baseInZip); err != nil {
		return "", err
	}

	return fullZipPath, nil
}

func (z *zipper) addFilesToZip(w *zip.Writer, basePath, baseInZip string) error {
	files, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		fullFilepath := filepath.Join(basePath, file.Name())
		z.logger.Trace("adding item to zip", z.logger.Args(
			"path", fullFilepath,
			"baseInZip", baseInZip,
		))

		if file.IsDir() {
			if err := z.addFilesToZip(w, fullFilepath, filepath.Join(baseInZip, file.Name())); err != nil {
				return err
			}
		} else if file.Type().IsRegular() {
			dat, err := os.ReadFile(fullFilepath)
			if err != nil {
				return err
			}
			f, err := w.Create(filepath.Join(baseInZip, file.Name()))
			if err != nil {
				return err
			}
			_, err = f.Write(dat)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
