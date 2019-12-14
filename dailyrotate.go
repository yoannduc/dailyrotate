// Copyright 2019 Yoann Duc. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Package dailyrotate is a daily rotating file writer
//
// The rotated file will be of format YYYY-MM-DD-basename
//
// A trivial example is:
//
// package main
//
// import "github.com/yoannduc/dailyrotate"
//
// func main() {
// 	rf, err := dailyrotate.New("/tmp/testfile.log", 3)
// 	if err != nil {
// 		// handle your error
// 	}
//
// 	rf.RotateWrite()
// }
//
package dailyrotate

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultFilePath Default file path to save file when using NewWithDefaults
	DefaultFilePath = "/tmp/rotating.log"
	// DefaultMaxAge Default max age to keep files when using NewWithDefaults
	DefaultMaxAge = 7
	// fileFlag Flag to open the files with
	fileFlag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	// filePerm Permissions to open the files with
	filePerm = 0644
)

// RotateWriter Rotating writer object
type RotateWriter struct {
	lock sync.Mutex
	// file Represents an open connection to the current day file
	file *os.File
	// time Represents the current time for writer to know if it should rotate
	time time.Time
	// FilePath Represents the filepath on which the rotating file
	// will be stored. Must be an absolute path
	FilePath string
	// MaxAge Represents the max number of file to keep before cleaning
	// after rotation. -1 for no cleaning
	MaxAge int
}

// New instanciate a new *RotateWriter with given path and max age
// Path must be an absolute path
// Max age is the number of old file to keep after rotation
func New(p string, ma int) (*RotateWriter, error) {
	p = filepath.Clean(p)
	if !filepath.IsAbs(p) {
		return nil, errors.New("Path should be absolute (\"" + p + "\" given)")
	}

	if ma < -1 {
		return nil, errors.New("MaxAge should be -1 or more (" +
			strconv.Itoa(ma) + " given)")
	}

	lf, err := os.OpenFile(
		p,
		fileFlag,
		filePerm,
	)
	if err != nil {
		return nil, err
	}

	return &RotateWriter{
		file:     lf,
		time:     time.Now(),
		FilePath: p,
		MaxAge:   ma,
	}, nil
}

// NewWithDefaults instanciate an new *RotateWriter with default path & max age
// Default path is "/tmp/rotating.log"
// Default max age is 7 days
func NewWithDefaults() (*RotateWriter, error) {
	return New(DefaultFilePath, DefaultMaxAge)
}

// Write satisfies the io.Writer interface.
func (rw *RotateWriter) Write(o []byte) (int, error) {
	rw.lock.Lock()
	defer rw.lock.Unlock()
	return rw.file.Write(o)
}

// RotateWrite performs a safe rotate and then write to file
func (rw *RotateWriter) RotateWrite(o []byte) (int, error) {
	err := rw.RotateSafe()
	if err != nil {
		return 0, err
	}

	return rw.Write(o)
}

// ShouldRotate returns a boolean indicating if file should rotate
// based on both today's date & file modified date compared to internal time
func (rw *RotateWriter) ShouldRotate() bool {
	ny, nm, nd := time.Now().Date()
	rwy, rwm, rwd := rw.time.Date()

	if ny != rwy || nm != rwm || nd != rwd {
		return true
	}

	if f, err := os.Stat(rw.FilePath); err == nil {
		fy, fm, fd := f.ModTime().Date()

		if fy != rwy || fm != rwm || fd != rwd {
			return true
		}
	}

	return false
}

// Rotate performs the rotation on the file. Verification on whether
// the file should rotate or not should be performed before calling Rotate.
// Rotate rename the current file in format YYYY-MM-DD_filename and then
// open a new file and then clean old file up to max age if max age is not -1
func (rw *RotateWriter) Rotate() error {
	rw.lock.Lock()
	defer rw.lock.Unlock()

	// Close existing file if open
	if rw.file != nil {
		err := rw.file.Close()
		if err != nil {
			return err
		}
		rw.file = nil
	}

	d := filepath.Dir(rw.FilePath)
	b := filepath.Base(rw.FilePath)
	if lf, err := os.Stat(rw.FilePath); err == nil {
		err = os.Rename(
			rw.FilePath,
			d+"/"+lf.ModTime().Format("2006-01-02")+"-"+b,
		)
		if err != nil {
			return err
		}
	}

	f, err := os.OpenFile(
		rw.FilePath,
		fileFlag,
		filePerm,
	)
	if err != nil {
		return err
	}
	rw.file = f

	// Clean only if MaxAge != -1. -1 is keep forever
	if rw.MaxAge > -1 {
		if err = rw.cleanOldFiles(); err != nil {
			return err
		}
	}

	rw.time = time.Now()
	return nil
}

// RotateSafe internally uses ShouldRotate and then Rotate if needed
func (rw *RotateWriter) RotateSafe() error {
	if rw.ShouldRotate() {
		return rw.Rotate()
	}

	return nil
}

// cleanOldFiles cleans old files up to max age
func (rw *RotateWriter) cleanOldFiles() error {
	dir := filepath.Dir(rw.FilePath)
	bname := filepath.Base(rw.FilePath)
	// aut is our list of authorized files that will not be removed
	aut := make(map[string]struct{})

	// Populate the list of authorized files based on file name and date
	// of previous days up to MaxAge days
	for i := 1; i <= rw.MaxAge; i++ {
		aut[time.Now().AddDate(0, 0, -i).Format("2006-01-02")+"-"+bname] = struct{}{}
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		fn := f.Name()
		// Do not remove if file begin with ".",
		// ends with basename or is basename (is current file)
		if strings.HasPrefix(fn, ".") ||
			fn == bname ||
			!strings.HasSuffix(fn, bname) {
			continue
		}

		// Do not remove if file is authorized
		_, ok := aut[fn]
		if ok {
			continue
		}

		err = os.Remove(dir + "/" + fn)
		if err != nil {
			return err
		}
	}

	return nil
}
