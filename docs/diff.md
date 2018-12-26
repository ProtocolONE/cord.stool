# Diff command
   Generate the difference between two directories

## Usage
```sh
   cord.stool.exe diff [command options] [arguments...]
```

## Options
```sh
   --old value, -o value    Path to old files
   --new value, -n value    Path to new files
   --patch value, -p value  Path to patch files
```

## Description
   Command **diff** generates the difference between two directories. It is using VCDIFF/RFC 3284 streams provided [Xdelta](http://xdelta.org/) library.</br>
   Use option **--old** to specify path to directory contains old files and option **--new** to specify path to directory contains  new files. The application generates the difference between files and creates patch in folder specified by option **--patch**.