/*
zfsver
------

zfsver takes a single argument, which must be a path to a regular
file on a ZFS filesystem. The program will search for other copies
of the file in snapshots inside the same dataset, reporting to you
what it finds.

With the `-v` flag, the modification time, size (in bytes) and full
path to each snapshot file is displayed. Without it, you get a
one-line summary.

This is the first thing I have written in Go. It is probably very bad.
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func die(msg string, exit_code int) {
	fmt.Println(msg)
	os.Exit(exit_code)
}

func snapdir(dir string) (string, string, error) {

	// Walk up the filesystem looking for a .zfs directory. Returns
	// the base of the ZFS dataset, the snapshot directory, and an
	// error. I thought it would be better to use `df` to get the root
	// of the filesystem holding the file, and use that, but it
	// doesn't work for loopback mounted filesystems in local zones.

	for {
		s_dir := filepath.Join(dir, ".zfs", "snapshot")
		_, err := os.Stat(s_dir)

		if err == nil {
			return dir, s_dir, nil
		}

		if filepath.Dir(dir) == dir {
			return "", "", fmt.Errorf("could not find snapdir")
		}

		dir = filepath.Dir(dir)
	}
}

func search_snaps(snap_dir string, relative_path string,
	snaplist *[]string) []string {

	// Look for the file in a list of snapshots, and return a slice of
	// stat() info about it if it's there.

	var matches []string

	for _, snap := range *snaplist {
		f := filepath.Join(snap_dir, snap, relative_path)
		st, err := os.Stat(f)

		if err == nil {
			matches = append(matches, fmt.Sprintf("%s %v %s\n",
				st.ModTime().Format("2006-01-02 15:04:05"), st.Size(), f))
		}
	}

	return matches
}

func count_unique_files(matches *[]string) int {

	// Given a pointer to a list of matches (from search_snaps()),
	// return the number of "unique" files, based on a combination of
	// timestamp and path.

	found := make(map[string]bool)

	for _, raw := range *matches {
		desc := strings.Join(strings.Fields(raw)[1:2], "")

		if !found[desc] {
			found[desc] = true
		}
	}

	return len(found)
}

func display_list(matches *[]string) {

	// Simple display function

	sort.Strings(*matches)
	for _, match := range *matches {
		fmt.Printf(match)
	}
}

func main() {
	// The user gives us a file. Probably in the cwd, but possibly
	// not. Do the usual sanity checks.

	verbosePtr := flag.Bool("v", false, "show all versions of file")
	flag.Parse()

	if len(flag.Args()) != 1 {
		die("please supply a single filename", 1)
	}

	f_path, err := filepath.Abs(flag.Args()[0])

	if err != nil {
		die("cannot resolve file path", 2)
	}

	f_path, _ = filepath.EvalSymlinks(f_path)

	if si, err := os.Stat(f_path); os.IsNotExist(err) {
		die(fmt.Sprintf("Cannot find %s", f_path), 3)
	} else if !si.Mode().IsRegular() {
		die(fmt.Sprintf("%s is not a regular file\n", f_path), 3)
	}

	// Walk up the directory hierarchy, looking for a ZFS snapshot
	// directory.

	fs_root, snap_dir, err := snapdir(f_path)

	if err != nil {
		die("File is not on a ZFS filesystem.", 7)
	}

	relative_path, err := filepath.Rel(fs_root, f_path)

	if err != nil {
		die("Could not find relative path.", 8)
	}

	// For every snapshot, look to see if our file exists. If it does,
	// report it.

	d, err := os.Open(snap_dir)

	if err != nil {
		die("Cannot read snapshot directory.", 5)
	}

	snaplist, err := d.Readdirnames(-1)

	if err != nil {
		die("Cannot read snapshots.", 5)
	}

	matches := search_snaps(snap_dir, relative_path, &snaplist)

	if len(matches) == 0 {
		fmt.Println("file not found in any snapshots")
		os.Exit(0)
	}

	uniques := count_unique_files(&matches)

	if *verbosePtr == false {
		fmt.Printf("found %d versions of file in %d snapshots.\n",
			uniques, len(snaplist))
	} else {
		display_list(&matches)
	}
}
