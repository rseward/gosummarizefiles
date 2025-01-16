# summarizefiles

A simple utility to summarize files by size:

---
    2024-07-18 22:31:56 /home   min mdate:  1980-01-31   max mdate:  2024-07-08 scanned:   8.6G   
    ⠴
    |          a:       2.8G in 2185 fi|        pak:      17.5M in 1 files
    |         so:    1018.7M in 1622 fi|     vlpset:      17.1M in 19 file
    |       html:     525.3M in 43170 f|         ps:      16.8M in 7 files
    |   snapshot:     411.0M in 12 file|       3ssl:      16.2M in 20748 f
    |          1:     398.7M in 1318 fi|        zig:      16.0M in 892 fil
    |          h:     265.6M in 23087 f|        exe:      14.9M in 124 fil
    |         py:     248.0M in 25525 f|         db:      14.4M in 21 file
---

by time modified:

---
    2024-07-18 22:33:29 /home   min mdate:  1980-01-31   max mdate:  2024-07-08 scanned:  12.4G   
    ⠲
    |2024-07-08:       2.6G in 17323 files      
    |2024-07-01:     227.3K in 2 files          
    |2024-06-28:      17.1M in 7 files          
    |2024-06-26:       1.7M in 171 files        
    |2024-06-19:       2.5M in 92 files         
    |   2024-06:       1.9G in 35141 files      
    |   2024-05:     801.0M in 28183 files      
    |   2024-04:     233.1M in 12998 files
---

by lines of text:

---
    2024-07-18 22:36:01 ..src/play   min mdate:  2023-07-29   max mdate:  2024-07-18 scanned: 107.5M
    ⠴
    |        js:     116002 lines in 965 files  |       sh~:         25 lines in 5 files    
    |        ts:     114782 lines in 289 files  |   feature:         25 lines in 1 files    
    |        md:      12027 lines in 213 files  |      toml:         24 lines in 1 files    
    |        cs:      10830 lines in 203 files  |       sql:         22 lines in 4 files    
    |       mjs:      10162 lines in 16 files   |        el:         19 lines in 1 files    
    |         c:       8262 lines in 13 files   |       go~:         18 lines in 1 files    
    |       cjs:       6789 lines in 16 files   |       pyi:         17 lines in 1 files
---

I personally use the tool to verify the transfer of files after rsync or the restore of data
from a backup. The tool is also useful for observing recent modifications to a directory tree.

## Prereqs for building and running

### Ubuntu

- apt-get install libmagic-dev -y

### Fedora

- libmagic-devel

## Related Projects

A slightly faster version of this utility is written in C. If a harder to build but faster to use utility is your thing please head there.

- https://github.com/rseward/summarizefiles

## Building the project

This version of the project uses the go bindings for libmagic. This design choice leads to 
dynamically linked builds to link to libmagic. 

A future version may replace that binding with a pure go implementation using heurestics based on the file extensions. When this is complete, it should allow for a static executable (more portable) to be built. 

### Build

---
    make build
---


