### Basics

If you don't want to bother with the whole window system, look into `basic.go`
file which contains utility functions to move cursor around, output text into
the terminal and turn text attributes on or off.

### Input capture

`tui.StartListening` launches a goroutine that sends the input into `tui.Input`
channel.

### Window system overview

Base interface is Drawable. It represents a rectangle that can draw on the
screen. Drawables can contain other Drawables, and each child Drawable has
a position and a size within its parent Drawable.

There's a root Drawable, Screen, which is returned when you initialize the
library with `tui.Init()`. Screen always occupies the whole terminal window.

When a Drawable is told to draw some text with something like
`drawable.MovePrintf`, it calculates a proper position within its parent
and calls parent's `MovePrintf` at given coordinates. Its parent does the
same thing, transforming the coordinates passed to the function into its
parent Drawable coordinate system, and so on. Eventually we will reach the
Screen drawable.

Screen has a double-buffered drawing system. When Screen is told to draw
something, it saves the input into a cell buffer. To flush the buffer, you
need to call its `Redraw` method. The method compares the current buffer state
with the previous buffer state, and then only redraws the parts that are
different, in the end replacing the "previous" state with the "current" one.

So to update some segment in some window, we always need to do two things in
succession:
- Call one of the drawing methods, like `window.MovePrint` or `window.HLine`
- Call `window.Redraw()` or `screen.Redraw()`

This system allows to accumulate changes from several sub-windows and redraw
the screen only once, and gives you the control of when you want/need to
refresh your terminal window.

### Batteries not included

You need to implement your own window management system to utilize this
fully. This module gives you basic interface to draw and capture input,
it does not contain tools to maintain window hierarchy, window drawing order,
tracking currently active window, etc.

### Emojis, heiroglyphics, etc.

There's basic support for so called "wide" symbols. Don't expect the library
to be able to handle advanced ZWJ sequences and such.
