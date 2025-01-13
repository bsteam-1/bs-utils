package utils

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	fl "github.com/klauspost/compress/flate"
	gz "github.com/klauspost/compress/gzip"
	zl "github.com/klauspost/compress/zlib"

	"github.com/pkg/errors"
)

type ProgressFunc func(archivePath string)

func Unzip_Path(file, SaveFolder string, progressFunc ProgressFunc) error {
	Zip_data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return Unzip_buffer(Zip_data, SaveFolder, progressFunc)
}
func Unzip_buffer(file []byte, SaveFolder string, progressFunc ProgressFunc) error {
	Zip_bytes_reader := bytes.NewReader(file)
	Zip_Reader, err := zip.NewReader(Zip_bytes_reader, Zip_bytes_reader.Size())
	if err != nil {
		return err
	}
	Zip_Reader.RegisterDecompressor(zip.Deflate, func(r io.Reader) io.ReadCloser {
		return flate.NewReader(r)
	})
	for _, file := range Zip_Reader.File {
		file_path := strings.ReplaceAll(file.Name, "\\", "/")
		file_modifytime := file.Modified

		file_Check_backslash := strings.HasSuffix(file_path, "\\")
		file_info := file.FileInfo()
		file_save_folder := path.Join(SaveFolder, file_path)

		if progressFunc != nil {
			progressFunc(file_save_folder)
		}
		switch {
		case file_info.IsDir() || file_Check_backslash:
			if err := os.MkdirAll(file_save_folder, file_info.Mode()|os.ModeDir|100); err != nil {
				return errors.Errorf("Mkdir %v", err)
			}
		case file_info.Mode()&os.ModeSymlink != 0:
			f, err := file.Open()
			defer f.Close()
			if err != nil {
				return errors.Errorf("Symlink open %v", err)
			} else {
				name, err := ioutil.ReadAll(f)
				if err != nil {
					return errors.Errorf("Symlink file read %v", err)
				} else {
					if err = os.Symlink(file_save_folder, string(name)); err != nil {
						return errors.Errorf("Symlink create %v", err)
					} else {
						log.Println("symlink create", file_save_folder)
					}
				}
			}
		default:
			f, err := file.Open() // get io.readcloser for zip inside files
			defer f.Close()
			if err != nil {
				return errors.Errorf("zip inside file open error Detail :", err)
			} else {
				filebufs, err := ioutil.ReadAll(f) // io.readcloser to bytearray
				if err != nil {
					return errors.Errorf("file read error Detail :", err)
				} else {
					dirpath, _ := path.Split(file_save_folder) // get directory path (check if dir only)
					_, err := os.Stat(dirpath)
					if os.IsNotExist(err) { // if folder is not exist. make folder
						err = os.MkdirAll(dirpath, file_info.Mode()|os.ModeDir|100)
						if err != nil {
							return errors.Errorf("makedir error Detail :", err)
						} else {
							log.Println("folder create", dirpath)
						}
					}
					err = ioutil.WriteFile(file_save_folder, filebufs, file.Mode()) // save file
					if err != nil {
						return errors.Errorf("file create error Detail :", err)
					} else {
						err = os.Chtimes(file_save_folder, file_modifytime, file_modifytime) // edit for save file to file access time // modify time
						if err != nil {
							return errors.Errorf("file edit modify time error Detail :", err)
						}
					}
				}
			}
		}
		if file_path == "" {
			continue
		}
	}
	return nil
}
func Zip(FilesAndFolders []string, SaveZipPath string, Compresslevel int, progress ProgressFunc) error {
	zipfile, err := os.Create(SaveZipPath)
	defer zipfile.Close()
	if err != nil {
		return err
	}
	zipWriter := zip.NewWriter(zipfile)
	defer zipWriter.Close()

	for _, data := range FilesAndFolders {
		Path_default := filepath.Dir(data)

		err := filepath.Walk(data, func(pathi string, infoi os.FileInfo, erri error) error {
			if err != nil || infoi.IsDir() {
				return erri
			}

			Path_relative, err := filepath.Rel(Path_default, pathi)
			if err != nil {
				return err
			}

			Path_archive := path.Join(filepath.SplitList(Path_relative)...)

			if progress != nil {
				progress(Path_archive)
			}

			file, err := os.Open(pathi)
			defer file.Close()
			if err != nil {
				return err
			}
			zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (closer io.WriteCloser, e error) {
				compresslevel := flate.BestCompression
				switch Compresslevel {
				case 1:
					compresslevel = flate.NoCompression
				case 2:
					compresslevel = flate.DefaultCompression
				case 3:
					compresslevel = flate.BestSpeed
				case 4:
					compresslevel = flate.BestCompression
				default:
					compresslevel = flate.BestCompression
				}

				return flate.NewWriter(out, compresslevel)
			})
			zipfileWriter, err := zipWriter.CreateHeader(&zip.FileHeader{
				Name:     Path_archive,
				Method:   zip.Deflate,
				Modified: infoi.ModTime(),
				NonUTF8:  false,
			})
			if err != nil {
				return err
			}
			_, err = io.Copy(zipfileWriter, file)
			return err
		})
		if err != nil {
			return err
		}
	}
	return nil
}

type Gzip_Compress struct {
}
type Zlib_Compress struct {
}

func (g Gzip_Compress) Compress(str string) (string, error) {
	var b bytes.Buffer
	gz, _ := gz.NewWriterLevel(&b, fl.BestCompression)
	if _, err := gz.Write([]byte(str)); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	return b.String(), nil
}
func (g Gzip_Compress) Compress_BB(str []byte) ([]byte, error) {
	var b bytes.Buffer
	gz, _ := gz.NewWriterLevel(&b, fl.BestCompression)
	if _, err := gz.Write(str); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
func (g Gzip_Compress) Compress_B(str string) ([]byte, error) {
	var b bytes.Buffer
	gz, _ := gz.NewWriterLevel(&b, fl.BestCompression)
	if _, err := gz.Write([]byte(str)); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
func (g Gzip_Compress) Decompress(str string) (string, error) {
	gr, err := gz.NewReader(bytes.NewBuffer([]byte(str)))
	if err != nil {
		return "", err
	}
	defer func() {
		s := recover()
		if s != nil {
			log.Println(s)
		}
		gr.Close()
	}()
	data, err := ioutil.ReadAll(gr)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (g Gzip_Compress) Decompress_B(str []byte) ([]byte, error) {
	gr, err := gz.NewReader(bytes.NewBuffer([]byte(str)))
	defer gr.Close()
	data, err := ioutil.ReadAll(gr)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (z Zlib_Compress) Compress(str string) (string, error) {
	var b bytes.Buffer
	gz, _ := zl.NewWriterLevel(&b, fl.BestCompression)
	if _, err := gz.Write([]byte(str)); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (z Zlib_Compress) Decompress(str string) (string, error) {
	gr, err := zl.NewReader(bytes.NewBuffer([]byte(str)))
	defer gr.Close()
	data, err := ioutil.ReadAll(gr)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (z Zlib_Compress) Compress_B(str string) ([]byte, error) {
	var b bytes.Buffer
	gz, _ := zl.NewWriterLevel(&b, fl.BestCompression)
	if _, err := gz.Write([]byte(str)); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (z Zlib_Compress) Decompress_B(str []byte) ([]byte, error) {
	gr, err := zl.NewReader(bytes.NewBuffer(str))
	defer gr.Close()
	data, err := ioutil.ReadAll(gr)
	if err != nil {
		return nil, err
	}
	return data, nil
}
