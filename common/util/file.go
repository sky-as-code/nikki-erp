package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	pathUtil "path/filepath"
	"strings"
)

const DEFAULT_FILE_PERMISSION = 0664
const DEFAULT_DIRECTORY_PERMISSION = 0775

type ApiError struct {
	Code  int    `json:"code,omitempty"`
	Error string `json:"error"`
}

func NewApiError(text string) *ApiError {
	return &ApiError{Error: text}
}
func NewApiCodeError(text string, code int) *ApiError {
	return &ApiError{Error: text, Code: code}
}

func GetWorkFolder(jobId string, rootPath string, create bool) (string, error) {

	if len(jobId) < 10 {
		return "", fmt.Errorf("invalid job ID: %s", jobId)
	}

	folder := rootPath + "/" + jobId[0:3] + "/" + jobId[3:6] + "/" + jobId[6:8]
	err := os.MkdirAll(folder, 0775)
	if err != nil {
		return "", err
	}

	return folder, nil
}

func CheckFileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func GetFileSize(name string) (int64, error) {
	info, err := os.Stat(name)

	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

func CopyFile(infile, outfile string) error {

	reader, err := os.Open(infile)
	if err != nil {
		return err
	}
	defer reader.Close()

	writer, err := os.OpenFile(outfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, reader)
	if err != nil {
		return err
	}

	return nil
}

func WriteFile(filename string, data []byte) error {

	err := ioutil.WriteFile(filename, data, 0664)
	if err != nil {
		return err
	}
	return nil
}

func ReadFile(filename string) ([]byte, error) {

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func GetTempFileName(tempFolder string, pattern string) (string, error) {

	f, err := ioutil.TempFile(tempFolder, pattern)
	if err != nil {
		return "", err
	}
	tempFile := f.Name()
	f.Close()
	os.Remove(f.Name())

	return tempFile, nil
}

func ExecCommand(dir string, name string, flags ...string) (out []byte, err error) {
	cmd := exec.Command(name, flags...)
	if dir != "" {
		cmd.Dir = dir
	}
	log.Println("Executing command: ", strings.Join(cmd.Args, " "))
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Println("Error: ", err.Error())
		log.Println("Output: ", string(out))
		return out, err
	}
	return out, nil
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func GetFileName(file string) string {

	if !strings.Contains(file, "/") {
		return file
	}
	parts := strings.Split(file, "/")
	return parts[len(parts)-1]
}

func NewNameSameExt(oldFilePathWithExt string, newFileNameNoExt string) string {
	ext := pathUtil.Ext(pathUtil.Base(oldFilePathWithExt))
	newFileNameSameExt := fmt.Sprintf("%s.%s", newFileNameNoExt, ext)
	return newFileNameSameExt
}
