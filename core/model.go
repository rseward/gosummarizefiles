// core package contains all the components needed by the summarizefiles utility.
package core

import (
	"fmt"
	"os"
	"sort"
	"time"
)

const (
	DateOnly = "2006-01-02"
)

// ProgramOpts type is used to pass the CLI and console size parameters around to the various components.
type ProgramOpts struct {
	Log     bool
	Ext     bool
	Time    bool
	Debug   bool
	Lines   bool
	ConCols int
	ConRows int
}

// SummaryEntry type represents the summary information for a group of files collesced together because of a
//
//	shared attribute (extension, time period modified, etc.).
type SummaryEntry struct {
	Group      string
	Label      string
	TotalBytes uint64
	LineCount  int
	FileCount  int32
	MinModTime time.Time
	MaxModTime time.Time
	Display    string
}

type SummaryEntryMap map[string]SummaryEntry

// FileSummary type represents the summary information for all files scanned.
type FileSummary struct {
	Root           string
	RootDisplay    string
	Total          uint64
	MaxModTime     time.Time
	MinModTime     time.Time
	Entries        SummaryEntryMap
	Groups         GroupMap
	ExceptionCount int
}

// NewFileSummary construct a FileSummary instance.
func NewFileSummary(root string) FileSummary {
	summ := FileSummary{}
	//m := make(map[string]int64)
	summ.Entries = NewSummaryEntryMap()
	summ.Groups = NewGroupMap()

	return summ
}

// SummaryGroup type is a collection of entries that should be grouped together for display / sorting purposes.
type SummaryGroup struct {
	Name    string
	Entries SummaryEntryMap
}

type EntryList []SummaryEntry
type GroupMap map[string]SummaryGroup

func (el EntryList) Len() int           { return len(el) }
func (el EntryList) Less(i, j int) bool { return el[i].TotalBytes < el[j].TotalBytes }
func (el EntryList) Swap(i, j int)      { el[i], el[j] = el[j], el[i] }

// TODO: remove this!
// TODO: Is there a better way to sort by lines instead of bytes by creating a new sub class. Gah!
// type EntryListByLines EntryList

// func (el EntryListByLines) Len() int           { return len(el) }
// func (el EntryListByLines) Less(i, j int) bool { return el[i].LineCount < el[j].LineCount }
// func (el EntryListByLines) Swap(i, j int)      { el[i], el[j] = el[j], el[i] }

// NewEntryList construct a EntryList instance.
func NewEntryList(size int) EntryList {
	return make(EntryList, size)
}

// NewSummaryEntryMap construct a Map of SummaryEntry instances.
func NewSummaryEntryMap() SummaryEntryMap {
	return make(SummaryEntryMap, 100)
}

// NewSummaryEntry construct a SummaryEntry instance.
func NewSummaryEntry() SummaryEntry {
	ret := SummaryEntry{}
	return ret
}

// NewGroupMap construct a GroupMap instance.
func NewGroupMap() GroupMap {
	return make(GroupMap, 100)
}

// AddEntry adds or updates a file entry into a SummaryEntryMap. Returns the entry created or updated.
func (semap SummaryEntryMap) AddEntry(popts *ProgramOpts, fs *FileSummary, label string, path string, finfo os.FileInfo) SummaryEntry {
	// Given a Map of SummaryEntries add or update one
	fsize := finfo.Size()
	se, ok := semap[label]
	if !ok {
		se = NewSummaryEntry()
		se.TotalBytes = 0
		se.Label = label
		se.FileCount = 0
		se.MaxModTime = finfo.ModTime()
		se.MinModTime = finfo.ModTime()
	}
	se.Label = label
	se.TotalBytes += uint64(fsize)
	se.FileCount++
	if finfo.ModTime().After(se.MaxModTime) {
		se.MaxModTime = finfo.ModTime()
	}
	if finfo.ModTime().Before(se.MinModTime) {
		se.MinModTime = finfo.ModTime()
	}
	semap[label] = se
	if popts.Lines {
		lc, err := CountLines(fs, path)
		if err != nil {
			fs.ExceptionCount += 1
			panic(err)
		} else {
			se.LineCount += lc
		}
		if popts.Debug {
			fmt.Printf("%s: lines = %d\n", finfo.Name(), se.LineCount)
		}
	}

	return se
}

// TODO: Should I stay or should I go?
/*
func (semap SummaryEntryMap) GetEntryList() EntryList {
	el := NewEntryList(len(semap))

	// iterate over the map values to create an array list

	return el

}
*/

// NewSummaryGroup construct a SummaryGroup instance.
func NewSummaryGroup(name string) SummaryGroup {
	ret := SummaryGroup{}
	ret.Name = name
	ret.Entries = NewSummaryEntryMap()
	return ret
}

// AddEntryByExt add or update a file entry. Summarize by file extension.
func (fs *FileSummary) AddEntryByExt(popts *ProgramOpts, fext string, path string, finfo os.FileInfo) SummaryEntry {
	fsize := finfo.Size()
	se := fs.Entries.AddEntry(popts, fs, fext, path, finfo)

	if fs.MaxModTime.IsZero() || finfo.ModTime().After(fs.MaxModTime) {
		fs.MaxModTime = finfo.ModTime()
	}
	if fs.MinModTime.IsZero() || finfo.ModTime().Before(fs.MinModTime) {
		fs.MinModTime = finfo.ModTime()
	}
	fs.Entries[fext] = se
	fs.Total += uint64(fsize)
	if popts.Lines {
		lc, err := CountLines(fs, path)
		if err != nil {
			fs.ExceptionCount += 1
			panic(err)
		} else {
			se.LineCount += lc
		}
		if popts.Debug {
			fmt.Printf("%s: lines = %d\n", finfo.Name(), se.LineCount)
		}
	}

	return se
}

// AddEntryByTime add or update a file entry. Summarize by time period file was modified. Return the entry.
func (fs *FileSummary) AddEntryByTime(popts *ProgramOpts, path string, finfo os.FileInfo) SummaryEntry {
	// Pass
	group, label := GetTimeGroup(finfo)
	//fmt.Printf("%v: %v, %v\n", finfo.Name(), group, label)
	fsize := finfo.Size()
	se := fs.Groups.AddEntry(popts, fs, group, label, path, finfo)
	se.Label = label
	if finfo.ModTime().After(se.MaxModTime) {
		se.MaxModTime = finfo.ModTime()
	}
	if finfo.ModTime().Before(se.MinModTime) {
		se.MinModTime = finfo.ModTime()
	}
	if fs.MaxModTime.IsZero() || finfo.ModTime().After(fs.MaxModTime) {
		fs.MaxModTime = finfo.ModTime()
	}
	if fs.MinModTime.IsZero() || finfo.ModTime().Before(fs.MinModTime) {
		fs.MinModTime = finfo.ModTime()
	}
	fs.Total += uint64(fsize)
	if popts.Lines {
		lc, err := CountLines(fs, path)
		if err != nil {
			fs.ExceptionCount += 1
			panic(err)
		} else {
			se.LineCount += lc
		}
	}

	//fmt.Printf("%+v\n", se)
	//fmt.Printf("%+v\n", fs.Groups)
	//fmt.Printf("fs.Total=%+v\n", fs.Total)

	return se

}

// GetTimeGroup determines time group and label for a file. Group is a broad grouping of the files 'less than a month', 'less than a year', 'older'.
// Label is something like YYYYMMDD
func GetTimeGroup(finfo os.FileInfo) (string, string) {
	modtime := finfo.ModTime()
	now := time.Now()
	group := ""
	label := modtime.Format(DateOnly)

	age := int(now.Sub(modtime).Hours() / 24.0)
	//fmt.Printf("modtime=%v ageindays=%d\n", modtime, age)

	if age < 30 {
		group = "01month"
	} else if age < 365 {
		group = "02year"
		label = label[0:7]
	} else {
		group = "03older"
		label = label[0:4]
	}

	return group, label
}

// TODO: Should I stay or should I go?
/*
func (sg *SummaryGroup) AddEntry(popts *ProgramOpts, fs *FileSummary, label string, path string, finfo os.FileInfo) SummaryEntry {
	se := sg.Entries.AddEntry(popts, fs, label, path, finfo)
	return se
}
*/

// AddEntry method for GroupMap objects. Add or update a file entry to the specified group and label. Return the entry.
func (gm GroupMap) AddEntry(popts *ProgramOpts, fs *FileSummary, group string, label string, path string, finfo os.FileInfo) SummaryEntry {
	sg, ok := gm[group]
	if !ok {
		sg = NewSummaryGroup(group)
	}
	se := sg.Entries.AddEntry(popts, fs, label, path, finfo)
	//sg.Entry.TotalBytes += uint64(finfo.Size())
	//sg.Entry.FileCount++
	gm[group] = sg
	//fmt.Printf("GroupMap.AddEntry %d: %d lines in %d files\n", se.Label, se.LineCount, se.FileCount)
	return se
}

// AddEntryToGroup method for FileSummary objects. Add or update a file entry based on group and label membership. Return the entry.
func (fs *FileSummary) AddEntryToGroup(popts *ProgramOpts, group string, label string, path string, finfo os.FileInfo) SummaryEntry {
	return fs.Groups.AddEntry(popts, fs, group, label, path, finfo)
}

// SortEntriesByBytes given a map of entries sort them by bytes.
func SortEntriesByBytes(summ map[string]SummaryEntry) EntryList {

	el := make(EntryList, 0, len(summ))
	for _, entry := range summ {
		el = append(el, entry)
	}

	sort.Sort(sort.Reverse(el))

	return el
}

// SortEntriesByLines given a map of entries sort them by line count.
func SortEntriesByLines(summ map[string]SummaryEntry) EntryList {
	el := make(EntryList, 0, len(summ))
	for _, entry := range summ {
		el = append(el, entry)
	}

	sort.Slice(el, func(i, j int) bool {
		return el[i].LineCount > el[j].LineCount
	})

	return el
}

// Calculate an appropriate RootPath for display taking into consideration the terminal size
func (self *FileSummary) SetDisplayRootPath(opts *ProgramOpts) {
	fmt.Printf("SetDisplayRootPath ConCols=%d", opts.ConCols)
	rootpathlen := len(self.Root)
	if opts.ConCols <= 97 || rootpathlen+95 < opts.ConCols {
		// Plenty of room to display the root in it's entirety
		self.RootDisplay = self.Root
	} else {
		// rootpath needs to be shortened
		rootpathlimit := opts.ConCols - 97 + 5
		if rootpathlen < rootpathlimit {
			// Show in entirety. Unreachable?
			self.RootDisplay = self.Root
		} else {
			self.RootDisplay = ".." + self.Root[rootpathlen-rootpathlimit:]
		}
	}
}
