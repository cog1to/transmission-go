# Development guide

This is a short run-down of app modules and how they work

## Main modules

- utils: Basic helper functions for working with slices, TTYs and such.
- worker: Another utility module that is mostly used to wrap repeating pieces of
 code.
- transmission: Wrapper for Transmission daemon protocol.
- logger: A primitive logging system to output debug logs into a file. Useful
 since the app itself manipulates and writes into the TTY in "curses"-like mode.
- suggestions: Directory prediction/suggestion logic based on input incomplete
 path string.
- tui: "curses"-like drawing functions, and some control functions to switch
 terminal modes
- list: TUI/window module dedicated to formatting and outputting list data with
 additional logic like scroll area support and selection support.
- windows: Main "windows" of the app, like torrent list, torrent detail,
 input dialogs, etc. This is the core of the application logic, including a
 `manager` object that controls the window "stack".

## Core loop

The `main.go` file does the initial setup:
  - initializes the TUI mode for the TTY 
  - creates an Transmission daemon client
  - creates a window manager with a root torrent list window
  - start the manager's event loop

The event loop of the WindowManager is pretty simple (see
`func (manager *WindowManager) Start()`). It listens to keypresses
from the TTY and redirects them to the "top" of the window stack. It also
listens to the 'draw' events of the top window and triggers terminal window
redraws if needed.

## Windows

The windows are responsible for updating their content (usually via Transmission
client calls) and responding to user inputs. Windows usually have `OnInput`
function that handles user input and some `Worker`s to poll the data from the
daemon. The majority of the logic lives in two main windows: `list.go`, which is
responsible for displaying torrent list as reported by the client, and
`details.go`, which displays torrent details, e.g. file list with download
status, current total progress and so on.

There are a couple of generic helper windows, like `prompt.go`, which is used
to ask user for some input, like setting donwload/upload speed limits, and
`cheatsheet.go`, which just displayes a "pretty" formatted list of key bindings
for the window.

There's also a generalized `input_field.go` struct that wraps some common logic
for displaying interactable fields. It is used in `Prompt` windows, and also in
`add_torrent.go` for getting torrent link/file and output directory.

The input handing/displaying logic is not so straightforward, so take some time
to read the code in `AddTorrentWindow`. Particularly pay attention how it
interacts with the window manager to add/remove InputReaders from the stack when
changing focus from one field to another.

## Drawing

Drawing is done in a pretty simple diff-like fashion. At the top of the window
manager, there's a Screen window that maintains two display buffers: one is
currently drawn on the screen, and another one "accumulates" all of the Draw 
calls from windows. When it's time to "flush" the buffer, it gathers the diff
between two buffers and updates terminal characters/cells that were affected
in the "display buffer", while doing some terminal control-character calls to
actually redraw the terminal window. See `screen.go` and `window.go` files in
the `tui` module.