// vim:ts=4:sw=4:noexpandtab

// Indexing tool for Debian Code Search. The tool scans through an unpacked
// source package mirror and adds relevant files to the regular expression
// trigram index.
//
// One run takes about 3 hours when using a slow USB disk as TMPDIR and storing
// the index output on a SSD.
package main

import (
	"dcs/index"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var numShards = flag.Int("shards", 1,
	"Number of index shards (the index will be split into 'shard' different files)")
var mirrorPath = flag.String("mirrorPath",
	"/dcs-ssd/",
	"Where to place the index.<shard>.idx files in")
var unpackedPath = flag.String("unpackedPath",
	"/dcs-ssd/unpacked/",
	"Where to look for unpacked directories. Needs to have a trailing /")
var dry = flag.Bool("dryrun", false, "Don't write index files")

// Returns true when the file matches .[0-9]$ (cheaper than a regular
// expression).
func hasManpageSuffix(filename string) bool {
	return len(filename) > 2 &&
		filename[len(filename)-2] == '.' &&
		filename[len(filename)-1] >= '0' &&
		filename[len(filename)-1] <= '9'
}

func main() {
	flag.Parse()
	fmt.Println("Debian Code Search indexing tool")

	ix := make([]*index.IndexWriter, *numShards)

	if !*dry {
		for i := 0; i < *numShards; i++ {
			path := fmt.Sprintf("%s/index.%d.idx", *mirrorPath, i)
			ix[i] = index.Create(path)
			ix[i].Verbose = true
		}
	}

	skiplen := len(*unpackedPath)
	if (*unpackedPath)[len(*unpackedPath)-1] != '/' {
		skiplen += 1
	}

	// Walk through all the directories and add files matching our source file
	// regular expression to the index.

	cnt := 0
	filepath.Walk(*unpackedPath,
		func(path string, info os.FileInfo, err error) error {
			if dir, filename := filepath.Split(path); filename != "" {
				// Skip quilt’s .pc directories and "po" directories (localization)
				if info.IsDir() &&
					(filename == ".pc" ||
						filename == "po") {
					return filepath.SkipDir
				}

				// NB: we don’t skip "configure" since that might be a custom shell-script
				// Skip documentation, configuration files and patches.
				// NB: we actually skip some autotools files because they blow up our index otherwise
				// TODO: peek inside the files (we’d have to read them anyways) and check for messages that indicate that the file is generated. either by autoconf or by bison for example.
				if filename == "NEWS" ||
					filename == "COPYING" ||
					filename == "LICENSE" ||
					filename == "CHANGES" ||
					filename == "Makefile.in" ||
					filename == "ltmain.sh" ||
					filename == "config.guess" ||
					filename == "config.sub" ||
					filename == "depcomp" ||
					filename == "aclocal.m4" ||
					filename == "libtool.m4" ||
					strings.HasSuffix(filename, ".conf") ||
					// spell checking dictionaries
					strings.HasSuffix(filename, ".dic") ||
					strings.HasSuffix(filename, ".cfg") ||
					strings.HasSuffix(filename, ".man") ||
					strings.HasSuffix(filename, ".xml") ||
					strings.HasSuffix(filename, ".xsl") ||
					strings.HasSuffix(filename, ".html") ||
					strings.HasSuffix(filename, ".sgml") ||
					strings.HasSuffix(filename, ".pod") ||
					strings.HasSuffix(filename, ".po") ||
					strings.HasSuffix(filename, ".patch") ||
					strings.HasSuffix(filename, ".txt") ||
					strings.HasSuffix(filename, ".tex") ||
					strings.HasSuffix(filename, ".rtf") ||
					strings.HasSuffix(filename, ".docbook") ||
					strings.HasSuffix(filename, ".symbols") ||
					// Don’t match /debian/changelog or /debian/README, but
					// exclude changelog and readme files generally.
					(!strings.HasSuffix(dir, "/debian/") &&
						strings.HasPrefix(strings.ToLower(filename), "changelog") ||
						strings.HasPrefix(strings.ToLower(filename), "readme")) ||
					hasManpageSuffix(filename) {
					if *dry {
						log.Printf("skipping %s\n", filename)
					}
					return nil
				}
			}

			if info != nil && info.Mode()&os.ModeType == 0 {
				// We strip the unpacked directory path plus the following
				// slash, e.g. /dcs-ssd/unpacked plus /
				indexname := path[skiplen:]
				if *dry {
					log.Printf("adding %s as %s\n", path, indexname)
				} else {
					ix[cnt%*numShards].AddFile(path, indexname)
					cnt++
				}
			}
			return nil
		})
	if !*dry {
		for i := 0; i < *numShards; i++ {
			ix[i].Flush()
		}
	}
	os.Exit(0)
}
