# Diff command
   Generate the difference between files

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
   Command **diff** generates the difference between files.</br>
   Use option **--old** to specify path to old files and option **--new** to specify path to new files. The application generates the difference between files and creates patch in folder specified by option **--patch**.
