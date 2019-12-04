# Dailyrotate

Dailyrotate is a daily rotating file writer for go (golang). It is based on top of [os.OpenFile](https://golang.org/pkg/os/#OpenFile) and implements [io.Writer](https://golang.org/pkg/io/#Writer) interface. Dailyrotate also cleans the directory to always match the number of rotated files you want.

## Why

Some applications need rotating file, daily rotating files instead of weight based rotation, such as some web app logs.

## Usage

Let's start with a trivial example:

```go
package main

import "github.com/yoannduc/dailyrotate"

func main() {
    // Instanciate a new RotateWriter
    rf, err := dailyrotate.New("/tmp/myfile", -1)
    if err != nil {
        // Handle error your own way
    }

    // Check if RotateWriter should rotate
    if rf.ShouldRotate() {
        // Perform the rotation
        err = rf.Rotate()
        if err != nil {
            // Handle error your own way
        }
    }

    // Write to the RotateWriter
    rf.Write([]byte("Some data"))
}
```

Or with short conveignance methods:

```go
package main

import "github.com/yoannduc/dailyrotate"

func main() {
    // Instanciate a new RotateWriter with default params
    rf, err := dailyrotate.NewWithDefaults()
    if err != nil {
        // Handle error your own way
    }

    // Write to the RotateWriter
    // Rotate if needed before writing to the file
    err = rf.RotateWrite([]byte("Some data"))
    if err != nil {
        // Handle error your own way
    }
}
```

### Params

Instanciation is done with two params:

- **FilePath** a string indicationg path to the rotating file. Must be absolute.

- **MaxAge** an int indicating the number of rotated files to keep after rotation. To keep all files without limitation, must be -1. Note that MaxAge works with time, not with number of files, so passing 3 will keep last 3 days, not last 3 file.

### Default params

Dailyrotate can be instanciated with default params. If done so, **FilePath** wil be `"/tmp/rotating.log"` and **MaxAge** will be `7`.
