# Immersive Labs Leaderboard

This application makes use of the api.immersivelabs.online endpoints
found while looking at network requests which occured when a user
signs in and when the 'Leaderboard' navigation bar link is clicked.

> Note: This is not a web scraper

## Installation

Run this is the terminal:
```bash
git clone https://github.com/UTC-Reading/immersivelabs-leaderboard
cd immmersivelabs-leaderboard
go install
```

Or install the binaries from one of the
[releases](https://github.com/UTC-Reading/immersivelabs-leaderboard/releases)

## Usage
> To use this application you need an immersivelabs account

If installed using `go install`. Just run `immersivelabs-leaderboard` in
the terminal

If you downloaded the binary go to the directory of the binary and run
it there. Or add the path of the binary to the PATH environment variable
and run `immersivelabs-leaderboard` in the terminal.

The csv should be saved to $HOME/Downloads/ImmersiveLabs/*.csv for Unix
based systems or saved to %USERPROFILE%/Downloads/ImmersiveLabs/*.csv for
Windows
