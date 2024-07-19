/*
summarizefiles a command line utility to summarize the size of various groups of files

The utility current supports summaizing files by:

  - extensions

  - by date

    It can total file sizes by:

  - bytes

  - lines
    Usage:

    sf [flags] path

    The flags are:
    --help
    Show the help for the cli
    --log
    Output summary to a file after completion
    --ext
    Summarize by extension the default
    --time
    Summarize the files by the last modification date.
    --lines
    Summarize the file sizes of text files by their line count. (Requires libmagic)
    --debug
    Ra roh, something has gone wrong let's trace it!
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"summarizefiles/core"
	"time"
)

/*
usage: summarizefiles.py [-h] [--log] [--ext] [--year] [--day] [--days] [--time] [--debug] [--lines] N [N ...]

positional arguments:
  N            Directories to summarize

options:
  -h, --help   show this help message and exit
  --log, -l    Log output to file_summary.txt when finished.
  --ext, -e    Summarize files by extension
  --time, -t   Summarize files by date modified. Most sophiscated time summary. Try it!
  --debug, -v  Something don't work, time to debug!
  --lines, -L  Summarize text files by their line count
*/

func main() {
	var myopts core.ProgramOpts

	logPtr := flag.Bool("log", false, "Specify to log output to file_summary.txt")
	debugPtr := flag.Bool("debug", false, "Something don't work, time to debug!")
	extPtr := flag.Bool("ext", false, "Summarize files by extension")
	timePtr := flag.Bool("time", false, "Summarize files by date modified")
	linesPtr := flag.Bool("lines", false, "Summarize files line count")

	flag.Parse()

	myopts.Log = *logPtr
	myopts.Debug = *debugPtr
	myopts.Ext = *extPtr
	myopts.Time = *timePtr
	myopts.Lines = *linesPtr

	if flag.NArg() == 0 {
		fmt.Println("summarizefiles requires a directory to examine!")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("Summarizing Files now...")
	_ = SummarizeFiles(flag.Arg(0), &myopts)

}

// SummarizeFile summarizes a file by program options.
func SummarizeFile(popts *core.ProgramOpts, summ *core.FileSummary, path string, info os.FileInfo) {

	if popts.Time {
		SummarizeFileByTime(popts, summ, path, info)
		return
	}
	SummarizeFileByExt(popts, summ, path, info)
}

// SummarizeFileByTime summarizes a file by the time period it was modified.
func SummarizeFileByTime(popts *core.ProgramOpts, summ *core.FileSummary, path string, info os.FileInfo) {
	//group, label := core.GetTimeGroup(info)
	//fmt.Printf("%v, %v\n", group, label)
	summ.AddEntryByTime(popts, path, info)
}

// SummarizeFileByExt summarizes a file by it's extension.
func SummarizeFileByExt(popts *core.ProgramOpts, summ *core.FileSummary, path string, info os.FileInfo) {
	fcomps := strings.Split(path, ".")
	fext := "Other"
	lidx := len(fcomps) - 1
	if len(fcomps) > 1 {
		fext = fcomps[lidx]
	}
	if popts.Debug {
		fmt.Printf("%v: %d %v\n", fext, len(fcomps), fcomps)
	}
	// mark fext greater than 6 as unknown
	if len(fext) > 9 {
		fext = "Other"
	}

	if fext == "Other" {
		return
	}

	se := summ.AddEntryByExt(popts, fext, path, info)
	if popts.Debug {
		fmt.Printf("%s: %d lines in %d files\n", path, se.LineCount, se.FileCount)
	}
}

// SummarizeFiles main loop that drives scanning the files and summarizing them.
func SummarizeFiles(mydir string, myopts *core.ProgramOpts) error {
	fmt.Println(mydir)

	summ := core.NewFileSummary(mydir)
	summ.Root = mydir
	lastshow := time.Now()

	myopts.GetConsoleSize()
	summ.SetDisplayRootPath(myopts)

	core.ClearConsole(true)

	err := filepath.Walk(mydir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				SummarizeFile(myopts, &summ, path, info)
			}
			//fmt.Println(path, info.Size())
			if time.Since(lastshow).Milliseconds() > 300 {
				core.Show(myopts, &summ)
				lastshow = time.Now()
			}
			return nil
		})
	if err != nil {
		fmt.Println(err)
	}
	core.Show(myopts, &summ)
	if myopts.Log {
		core.Log(myopts, &summ)
	}

	return nil
}
