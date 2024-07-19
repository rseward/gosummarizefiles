// core package contains all the components needed by the summarizefiles utility.
package core

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func getConsoleSize() (width, height int, err error) {
	var termDim [4]uint16
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(0), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&termDim)), 0, 0, 0); err != 0 {
		return -1, -1, err
	}
	return int(termDim[1]), int(termDim[0]), nil
}

func ClearConsole(cls bool) {
	// TODO: explore using https://github.com/inancgumus/screen for cross platform
	if cls {
		fmt.Print("\033[2J")
	}
	fmt.Print("\033[H")
}

// FormatEntry formats an entry into a column width chunk of text for rendering into output window.
func FormatEntry(opts *ProgramOpts, entry SummaryEntry, colwidth int) string {

	var display string = ""
	if opts.Lines {
		display = fmt.Sprintf("%10s: %10v lines in %d files",
			entry.Label, entry.LineCount, entry.FileCount)
	} else {
		display = fmt.Sprintf("%10s: %10v in %d files",
			entry.Label, humansize(uint64(entry.TotalBytes)), entry.FileCount)
	}
	//display = strings.Repeat(" ", colwidth+1)

	if colwidth == -1 {
		return fmt.Sprintf("%s", display)
	} else {
		if len(display) > colwidth {
			return display[1:colwidth]
		}
	}
	return fmt.Sprintf("%-*s", colwidth, display)
}

// RenderGroups iterates over summary groups to include all entries with information useful to display.
func RenderGroups(popts *ProgramOpts, summ *FileSummary) EntryList {
	// Render files, first by their group and then their label
	keys := make([]string, 0, len(summ.Groups))
	for key := range summ.Groups {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	//gel := make(map[int]EntryList, len(summ.Groups))
	totalentries := 0

	for kidx := range keys {
		key := keys[kidx]
		group := summ.Groups[key]

		//gel[kidx] := RenderGroup(group)
		totalentries += len(group.Entries)
		//fmt.Printf("%s.totalentries=%d\n", key, totalentries)
	}
	//fmt.Printf("totalentries=%d\n", totalentries)

	// Combine all the group entries together
	allentries := make(EntryList, totalentries)
	allidx := 0
	for kidx := range keys {
		key := keys[kidx]
		group := summ.Groups[key]
		// TODO: sort the entries by their label in each group
		sorted := SortByLabels(group)

		for sidx := range sorted {
			allentries[allidx] = sorted[sidx]
			allidx++
		}
	}
	return allentries
}

// SortByLabels returns a list of entries sorted by their label in descending order.
func SortByLabels(group SummaryGroup) EntryList {
	var lmap map[string]SummaryEntry = make(map[string]SummaryEntry, len(group.Entries))
	var el EntryList = NewEntryList(len(group.Entries))

	for eidx := range group.Entries {
		lmap[group.Entries[eidx].Label] = group.Entries[eidx]
	}

	keys := make([]string, len(lmap))
	kidx := 0
	for key := range lmap {
		keys[kidx] = key
		kidx++
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	for kidx := range keys {
		key := keys[kidx]
		el[kidx] = lmap[key]
	}

	return el
}

var spinners string = "\u2832\u2834\u2826\u2816"
var tick int = 0

// Render drives the logic to render entries into columns for display while executing.
func Render(opts *ProgramOpts, summ *FileSummary) {
	colwidth := 35
	if opts.Time || opts.Lines {
		colwidth = 45
	}

	linedisp := make([]string, opts.ConRows)

	// timeline = "%20s%12s: %30s %12s: %30s bytes: %10d errs: %4d" % ( now, "min mdate", dispmindate, "max mdate", dispmaxdate, summ.TotalBytes, showexceptions )
	now := fmt.Sprintf("%v", time.Now())[0:19]
	dispmindate := fmt.Sprintf("%v", summ.MinModTime)[0:10]
	dispmaxdate := fmt.Sprintf("%v", summ.MaxModTime)[0:10]
	timeline := ""
	if opts.ConCols > 97 {
		timeline = fmt.Sprintf("%18s %s %11s: %11s %11s: %11s scanned: %6s errs: %3d", now, summ.RootDisplay, "min mdate", dispmindate, "max mdate", dispmaxdate, humansize(summ.Total), summ.ExceptionCount)
	} else {
		// A more compact status line for smaller terminals.
		timeline = fmt.Sprintf("%18s %s scanned: %6s errs: %3d", now, summ.RootDisplay, humansize(summ.Total), summ.ExceptionCount)
	}
	if len(timeline) > opts.ConCols {
		timeline = timeline[0:opts.ConCols] // truncate just to be sure
	}
	dcols := int(float64(opts.ConCols) / float64(colwidth))

	if (dcols * colwidth) > opts.ConCols {
		dcols -= 1
	}

	var el EntryList
	if opts.Time {
		el = RenderGroups(opts, summ)
	} else if opts.Lines {
		el = SortEntriesByLines(summ.Entries)
	} else {
		el = SortEntriesByBytes(summ.Entries)
	}

	if opts.Debug {
		fmt.Printf("len(el)=%v summ.TotalBytes=%d\n", len(el), summ.Total)
		fmt.Printf("Groups=%+v\n", summ.Groups)
	}

	// Render the list of entry display items into text columns into the linedisp array
	lineidx := 0
	colidx := 1
	for idx := 0; idx < len(el); idx++ {
		entry := el[idx]
		if opts.Lines {
			/*
				if idx == 0 {
					fmt.Printf("\n\n\n\n\n")
				}
				fmt.Printf("%s: %d lines in %d files\n", entry.Label, entry.LineCount, entry.FileCount) */
		}

		displayit := false

		if opts.Lines {
			if entry.LineCount > 0 {
				displayit = true
			}
		} else {
			if entry.TotalBytes > 1024 {
				displayit = true
			}
		}

		if displayit {
			var sb strings.Builder
			sb.WriteString(linedisp[lineidx])
			sb.WriteString(fmt.Sprintf("|%34s", FormatEntry(opts, entry, colwidth-2)))

			linedisp[lineidx] = sb.String()
			lineidx++

			if lineidx >= len(linedisp) {
				// We've reached the end of the column, move to the next
				lineidx = 0
				colidx += 1
				if colidx > dcols {
					break
				}
			}
		}
	}

	if !opts.Debug {
		ClearConsole(false)
	}
	fmt.Println(timeline)
	// fmt.Printf("%d %d %d\n", dcols, opts.ConCols, opts.ConRows)
	var sr rune = []rune(spinners)[tick%4]
	tick++
	//fmt.Println("\u2832")
	fmt.Println(string(sr))

	for idx := 0; idx < len(linedisp); idx++ {
		fmt.Println(linedisp[idx])
	}
}

// Log renders the final output into a file named: file_summary.txt
func Log(opts *ProgramOpts, summ *FileSummary) {
	var el EntryList
	if opts.Time {
		el = RenderGroups(opts, summ)
	} else {
		el = SortEntriesByBytes(summ.Entries)
	}

	if opts.Log {
		f, err := os.Create("file_summary.txt")
		if err != nil {
			panic(err)
		}
		outf := bufio.NewWriter(f)
		for idx := 0; idx < len(el); idx++ {
			entry := el[idx]
			outf.WriteString(fmt.Sprintf("%s\n", FormatEntry(opts, entry, -1)))
		}
		outf.Flush()
		f.Sync()
		f.Close()
		fmt.Println("Wrote summary to file_summary.txt\n")
	}
}

func (opts *ProgramOpts) GetConsoleSize() {
	if opts.ConCols == 0 {
		opts.ConCols, opts.ConRows, _ = getConsoleSize()
		opts.ConRows -= 3
		fmt.Printf("calculated: rows=%v cols=%v\n", opts.ConRows, opts.ConCols)
	}
}

// Show is called periodically to show the summary of files scanned so far.
func Show(opts *ProgramOpts, summ *FileSummary) {
	opts.GetConsoleSize()
	Render(opts, summ)
}

// humansize displays the size in bytes to a more human friendly output
func humansize(bytes uint64) string {
	gbytes := 1024 * 1024 * 1024
	mbytes := 1024 * 1024
	kbytes := 1024
	if bytes > uint64(gbytes) {
		g := float64(float64(bytes) / float64(gbytes))
		return fmt.Sprintf("%.1fG", g)
	} else if bytes > uint64(mbytes) {
		m := float64(float64(bytes) / float64(mbytes))
		return fmt.Sprintf("%.1fM", m)
	}
	k := float64(float64(bytes) / float64(kbytes))
	return fmt.Sprintf("%.1fK", k)
}

// main unit tests the getConsoleSize method
func main() {
	cols, rows, _ := getConsoleSize()
	fmt.Printf("rows: %v cols: %v\n", rows, cols)
}
