// core package contains all the components needed by the summarizefiles utility.
package core

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vimeo/go-magic/magic"
)

// initialized : has libmagic been initialized yet
var initialiazed bool = false

// debug : are we still trying to work out why something isn't working?
var debug bool = false

// CountLines give a file path, decide if the file is a text file and count the number of lines in the file.
func CountLines(summ *FileSummary, path string) (int, error) {
	if !initialiazed {
		magic.AddMagicDir(magic.GetDefaultDir())
	}

	mimetype := magic.MimeFromFile(path)
	//fmt.Printf("%s: %s\n", path, mimetype)
	if strings.Contains(mimetype, "text") {
		// TODO: count exceptions
		inf, err := os.Open(path)
		if err != nil {
			summ.ExceptionCount++
			return 0, err
		}

		lines, err2 := lineCounter(inf)
		if debug {
			fmt.Printf("%s: %d lines\n", path, lines)
		}
		if err2 != nil {
			summ.ExceptionCount++
			return 0, err2
		}
		return lines, nil
	} else {
		if debug {
			mylog := "lc_anomalies.txt"
			if len(mimetype) < 1 {
				f, err := os.OpenFile(mylog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					panic(err)
				}
				defer f.Close()
				outf := bufio.NewWriter(f)
				outf.WriteString(fmt.Sprintf("%s: mimetype=%s\n", path, mimetype))
				outf.Flush()
				f.Sync()
			}
		}
	}

	return 0, nil
}

// lineCounter do the core task of counting the number of lines in a text file.
func lineCounter(r io.Reader) (int, error) {
	// https://stackoverflow.com/questions/24562942/golang-how-do-i-determine-the-number-of-lines-in-a-file-efficiently
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
