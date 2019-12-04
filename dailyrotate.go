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

// TODO COMMENT GODOC

const (
	// DefaultFilePath Default file path to save file when using NewWithDefaults
	DefaultFilePath = "/tmp/rotating.log"
	// DefaultMaxAge Default max age to keep files when using NewWithDefaults
	DefaultMaxAge = 7
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
	// after rotation. 0 for no cleaning
	MaxAge int
}

func New(p string, ma int) (*RotateWriter, error) {
	p = filepath.Clean(p)
	if !filepath.IsAbs(p) {
		return nil, errors.New("Path should be absolute (\"" + p + "\" given)")
	}

	if ma < 0 {
		return nil, errors.New("MaxAge should be 0 or more (" +
			strconv.Itoa(ma) + " given)")
	}

	lf, err := os.OpenFile(
		p,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
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

func NewWithDefaults() (*RotateWriter, error) {
	lf, err := os.OpenFile(
		DefaultFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, err
	}

	return &RotateWriter{
		file:     lf,
		time:     time.Now(),
		FilePath: DefaultFilePath,
		MaxAge:   DefaultMaxAge,
	}, nil
}

// Write satisfies the io.Writer interface.
func (rw *RotateWriter) Write(o []byte) (int, error) {
	rw.lock.Lock()
	defer rw.lock.Unlock()
	return rw.file.Write(o)
}

func (rw *RotateWriter) RotateWrite(o []byte) (int, error) {
	err := rw.RotateSafe()
	if err != nil {
		return 0, err
	}

	return rw.Write(o)
}

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
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}
	rw.file = f

	// Clean only if MaxAge != 0. 0 is keep forever
	if rw.MaxAge > 0 {
		if err = rw.cleanOldFiles(); err != nil {
			return err
		}
	}

	rw.time = time.Now()
	return nil
}

func (rw *RotateWriter) RotateSafe() error {
	if rw.ShouldRotate() {
		return rw.Rotate()
	}

	return nil
}

func (rw *RotateWriter) cleanOldFiles() error {
	dir := filepath.Dir(rw.FilePath)
	bname := filepath.Base(rw.FilePath)
	// aut is our list of authorized files that will not be removed
	aut := make(map[string]struct{})

	// Populate the list of authorized files based on file name and date
	// of previous days up to MaxAge days
	for i := 0; i < rw.MaxAge; i++ {
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
