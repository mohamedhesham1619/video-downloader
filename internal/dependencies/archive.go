package dependencies

import (
	"archive/tar"
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"github.com/ulikunitz/xz"
)

// unzip extracts a zip file to dest directory
func unzipFile(src, dest string) error {
    r, err := zip.OpenReader(src)
    if err != nil {
        return err
    }
    defer r.Close()

    for _, f := range r.File {
        fpath := filepath.Join(dest, f.Name)

        if f.FileInfo().IsDir() {
            os.MkdirAll(fpath, 0755)
            continue
        }

        os.MkdirAll(filepath.Dir(fpath), 0755)

        rc, err := f.Open()
        if err != nil {
            return err
        }
        defer rc.Close()

        out, err := os.Create(fpath)
        if err != nil {
            return err
        }

        _, err = io.Copy(out, rc)
        out.Close()
        if err != nil {
            return err
        }
    }

    return nil
}

// untarXzFile extracts a .tar.xz file to dest directory
func untarXzFile(src, dest string) error {
    file, err := os.Open(src)
    if err != nil {
        return err
    }
    defer file.Close()

    xzReader, err := xz.NewReader(file)
    if err != nil {
        return err
    }

    tarReader := tar.NewReader(xzReader)

    for {
        header, err := tarReader.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }

        outPath := filepath.Join(dest, header.Name)

        switch header.Typeflag {
        case tar.TypeDir:
            os.MkdirAll(outPath, 0755)
        case tar.TypeReg:
            os.MkdirAll(filepath.Dir(outPath), 0755)
            outFile, err := os.Create(outPath)
            if err != nil {
                return err
            }
            io.Copy(outFile, tarReader)
            outFile.Close()
        }
    }

    return nil
}

// copyFile copies a file from src to dest
func copyFile(src, dest string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()

    out, err := os.Create(dest)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, in)
    return err
}