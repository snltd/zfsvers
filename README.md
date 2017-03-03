zfsver
------

`zfsver` takes a single argument, which must be a path to a regular
file on a ZFS filesystem. The program will search for other copies
of the file in snapshots inside the same dataset, reporting to you
what it finds.

With the `-v` flag, the modification time, size (in bytes) and full
path to each snapshot file is displayed. Without it, you get a
one-line summary.

This is the first thing I have written in Go. It is probably very bad.
