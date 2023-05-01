module github.com/cog1to/transmission-go

go 1.13

require (
    transmission v0.0.0
    tui v0.0.0
    windows v0.0.0
    logger v0.0.0
)

replace (
    transmission => ./transmission
    tui => ./tui
    windows => ./windows
    worker => ./worker
    utils => ./utils
    transform => ./transform
    list => ./list
    suggestions => ./suggestions
    logger => ./logger
)
