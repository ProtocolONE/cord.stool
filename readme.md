# Cord.Stool

[![Build Status](https://travis-ci.org/ProtocolONE/cord.stool.svg?branch=master)](https://travis-ci.org/ProtocolONE/cord.stool)
![License](https://img.shields.io/hexpm/l/plug.svg)

## Description
*Cord.Stool* is command-line tool that lets you:
 * Prepare game distributive for Cord.App distribution system 
 * Upload distributive or patch for different [cdn]()
 * Update download information in cord.api
 * Auto register patch in torrent tracker
 
## Prerequisites
 * [Go >=1.11.x](https://golang.org/dl/)
 
 
## Getting started
First install dependencies:
```sh
go get -u
```

If you want to specify version info and icon for Windows download `goversioninfo`
```sh
go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
```
  
To generate Windows resource file check [documentation](github.com/josephspurrier/goversioninfo/) or use default options:
```sh
go generate
```

Finally to build application run build command:
```sh
go build
```

## Usage

```sh
cord.stool.exe [global options] command [command options] [arguments...]
```

You can get more help for commands:
```sh
cord.stool.exe help [command]
```

Commands:
 * [create](docs\create.md)
 * [diff](docs\diff.md)
 * [push](docs\push.md)
 * [torrent](docs\torrent.md)
 * [upgrade](docs\upgrade.md)
 * [help](docs\create.md)