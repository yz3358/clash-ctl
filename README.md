# clash-ctl
A cli for controling clash server. It works well on Linux / MacOS. Not compatible with Powershell.

## Credit 
This project is just a copy of [Dreamacro's clash-ctl](https://github.com/Dreamacro/clash-ctl), with a few added features (that are inspired by [yichengchen's clashX](https://github.com/yichengchen/clashX)).

- proxies table sorted by delay 
- proxies benchmark test 
- proxy mode management

I think clashX is great, and I want to use it on linux, thus I created this project.

As of now, the program assumes the server is running with only 1 custom selector. This should work in most cases.

## Getting Started

```bash
go mod download
go build -o bin/clash-ctl

# run the binary
./bin/clash-ctl
```

## Key Features
```bash
# add a clash server (to your local clash controller, e.g. 127.0.0.1:9090) 
server add
# ... follow the steps to add a server 
# (config file stored as ${HOME}/.config/clash/ctl.toml)

server use <your_server_name>

# list proxies of the primary provider 
proxy ls 

# use the first one 
proxy use 0

# test proxy benchmark 
proxy bench

# show current proxy mode
mode

# set proxy mode to global
mode global
```
