package dailyrotate

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const (
	DefaultLogsFilePath = "/tmp/rotating.log"
	DefaultMaxAge       = 7
)

// RotateWriter ...
type RotateWriter struct {
	lock         sync.Mutex
	logsFile     *os.File
	time         time.Time
	LogsFilePath string `default:"/tmp/rotating.log"`
	MaxAge       int    `default:"7"`
}

func New(p string, ma int) (*RotateWriter, error) {
	p = filepath.Clean(p)
	if !filepath.IsAbs(p) {
		return nil, errors.New("Path should be absolute (\"" + p + "\" given)")
	}

	if ma < 1 {
		return nil, errors.New("MaxAge should be 1 or more (" +
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
		logsFile:     lf,
		time:         time.Now(),
		LogsFilePath: p,
		MaxAge:       ma,
	}, nil
}

func NewWithDefaults() (*RotateWriter, error) {
	lf, err := os.OpenFile(
		DefaultLogsFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, err
	}

	return &RotateWriter{
		logsFile:     lf,
		time:         time.Now(),
		LogsFilePath: DefaultLogsFilePath,
		MaxAge:       DefaultMaxAge,
	}, nil
}

// // Write satisfies the io.Writer interface.
// func (w *RotateWriter) Write(output []byte) (int, error) {
// 	w.lock.Lock()
// 	defer w.lock.Unlock()
// 	return w.logsFile.Write(output)
// }

// // Rotate perform the actual act of rotating and reopening file.
// func (w *RotateWriter) Rotate() (err error) {
// 	w.lock.Lock()
// 	defer w.lock.Unlock()

// 	if w.time.IsZero() {
// 		w.time = time.Now()
// 	} else if w.time.YearDay() == time.Now().YearDay() &&
// 		w.time.Year() == time.Now().Year() &&
// 		w.logsFile != nil {
// 		return
// 	}

// 	// Close existing file if open
// 	if w.logsFile != nil {
// 		err = w.logsFile.Close()
// 		if err != nil {
// 			return err
// 		}
// 		w.logsFile = nil
// 	}

// 	dirname := filepath.Dir(w.LogsFilePath)
// 	basename := filepath.Base(w.LogsFilePath)

// 	if logsFile, err := os.Stat(w.LogsFilePath); err == nil &&
// 		(w.time.YearDay() != logsFile.ModTime().YearDay() ||
// 			w.time.Year() != logsFile.ModTime().Year()) {
// 		err := os.Rename(
// 			w.LogsFilePath,
// 			dirname+"/"+logsFile.ModTime().Format("2006-01-02")+"-"+basename,
// 		)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	w.logsFile, err = os.OpenFile(
// 		w.LogsFilePath,
// 		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
// 		0644,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	kl := make(map[string]interface{})
// 	for i := 0; i < w.MaxAge; i++ {
// 		kl[time.Now().AddDate(0, 0, -i).Format("2006-01-02")+"-"+basename] = ""
// 	}

// 	files, err := ioutil.ReadDir(dirname)
// 	if err != nil {
// 		w.logsFile.Close()
// 		return err
// 	}

// 	for _, file := range files {
// 		fn := file.Name()
// 		if strings.HasPrefix(fn, ".") ||
// 			fn == basename ||
// 			!strings.HasSuffix(fn, basename) {
// 			continue
// 		}
// 		_, ok := kl[fn]
// 		if ok {
// 			continue
// 		}

// 		err = os.Remove(dirname + "/" + fn)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	w.time = time.Now()
// 	return
// }
