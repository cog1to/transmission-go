# transmission-go

Feature-poor ncurses-based `transmission-daemon` client.

## Reason

I wanted to learn some Go and do something useful in the process. Having an affinity for terminal and CLI tools, I figured this would be a nice project that I might be using on a daily basis myself.

## Features/screens

(For obvious reasons all torrent/filenames are obfuscated)

List of active torrents.

![list](img/list.png)
![adding torrent](img/new_torrent.png)

Torrent details.

![details](img/details.png)
![speed limit](img/down_limit_dialog.png)

## Usage

### Launch Arguments

`-h` -- hostname to connect to; default: `localhost`
`-p` -- port number; default: `9091`
`-o` -- obfuscate all torrent and filenames (added this just for making screenshots)

### Controls

##### List screen

| Keys  | Purpose |
|-------|---------|
| F1    | Show cheatsheet |
| q     | Exit |
| jk↑↓  | Move cursor up and down |
| l→    | Go to torrent details |
| Space | Toggle selection |
| c     | Clear selection |
| d     | Remove torrent(s) from the list (keep data) |
| D     | Delete torrent(s) along with the data |
| p     | Start/stop selected torrent(s) |
| L     | Set global download speed limit |
| U     | Set global upload speed limit |

##### Details screen

| Keys  | Purpose |
|-------|---------|
| F1    | Show cheatsheet |
| qh←   | Go back to torrent list |
| jk↑↓  | Move cursor up and down |
| Space | Toggle selection |
| c     | Clear selection |
| g     | Download/Don't download selected file(s) |
| p     | Change priority of selected file(s) |
| L     | Set torrent's download speed limit |
| U     | Set torrent's upload speed limit |

## Building

Obviously requires a working Go environment and headers/libs for ncurses. To properly support wide characters, I link against `ncursesw`

Building should be pretty straightforward:
```
go build
```

## Contributions

Feel free to use, modify, report bugs, create feature requests or pull requests.

## Acknowledgements

Wide-char support: [GeertJohan/cgo.wchar](https://github.com/GeertJohan/cgo.wchar)
Goncurses: [rthornton128/goncurses](https://github.com/rthornton128/goncurses). I had to modify it a little to expose wide-character functions of ncurses.

## License

GPLv3
