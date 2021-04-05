package main

import (
	"io/fs"
	"io/ioutil"
	"strconv"
)

type fileBackedInt struct {
	filename string
}

func newFileBackedInt(filename string) (fileBackedInt, error) {
	f := fileBackedInt{filename: filename}
	return f, f.write(0)
}

func (f fileBackedInt) read() (int, error) {
	data, err := ioutil.ReadFile(f.filename)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

func (f fileBackedInt) write(count int) error {
	data := []byte(strconv.Itoa(count))
	return ioutil.WriteFile(f.filename, data, fs.ModePerm)
}
