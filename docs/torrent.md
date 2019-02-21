# Torrent command
   Create torrent file and adding it to Torrent Tracker if needed it

## Usage
```sh
   cord.stool.exe torrent [command options] [arguments...]
```

## Options
```sh
   --source value, -s value           Path to game
   --file value, -f value             Path to torrent file
   --web-seeds value, --ws value      Slice of torrent web seeds
   --announce-list value, --al value  Slice of announce server url
   --piece-length value, --pl value   Torrent piece length (default: 512)
   --cord-url value                   Cord server url
   --cord-login value                 Cord user login
   --cord-password value              Cord user password
```

## Description
   Command **torrent** creates a torrent file.</br>
   Use option **--source** to specify path to source files and option **--file** to specify a created torrent file name. Option **--piece-length** specifies a value of Torrent pieces length (default: 512).</br>
   Use **--web-seeds** to specify a slice of torrent web seeds (multi option). Use **--announce-list** to specify a slice of announce server url (multi option).
   To add torrent file (**--file** value) to Torrent Tracker specify **--cord-url** value as Cord Server url, **--cord-login** value as Cord user login and **--cord-password** value as Cord user password
   
