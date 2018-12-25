# Torrent command
   Create torrent file

## Usage
```sh
   cord.stool.exe torrent [command options] [arguments...]
```

## Options
```sh
   --source value, -s value           Path to game
   --target value, -t value           Path for new torrent file
   --web-seeds value, --ws value      Slice of torrent web seeds
   --announce-list value, --al value  Slice of announce server url
   --piece-length value, --pl value   Torrent piece length (default: 512)
```

## Description
   Command **torrent** creates a torrent file.</br>
   Use option **--source** to specify path to source files and option **--target** to specify a created torrent file name.
