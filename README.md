# tvpkg — a friendly package manager for Termux

`tvpkg` (also available as `tvp`) is a wrapper around Termux's package
manager, supporting both **APT** and **Pacman**, with an interactive TUI and
a fast command-line interface.

Real help is the package that ships with Termux (`pkg`); tvpkg behaves like
it, but adds a fuzzy-search TUI and works with either backend.

## Install

tvpkg is packaged for both Termux package formats:

```sh
# APT
apt install tvpkg

# Pacman
pacman -S tvpkg
```

Or run it straight from source:

```sh
go build -trimpath -ldflags "-s -w" -o tvpkg .
./tvpkg
```

## Usage

Run without arguments to open the **TUI**:

```
tvpkg
```

Keybinds in the TUI:

| Key        | Action                                  |
|------------|-----------------------------------------|
| `↑/k`,`↓/j`| move through the package list           |
| `/`        | fuzzy-filter packages                   |
| `tab`      | cycle action (Install / Remove / Info)  |
| `enter`    | run the selected action                 |
| `q` / `esc`| quit / go back                          |

Non-interactive commands:

```
tvpkg install <pkgs...>   (i)  apt install -y / pacman -S
tvpkg remove  <pkgs...>   (r)  apt remove -y  / pacman -R
tvpkg search  <query>     (s)  apt-cache search / pacman -Ss
tvpkg update              (u)  apt update        / pacman -Sy
tvpkg list                (l)  list available packages
tvpkg list --installed    (li) list installed packages
tvpkg info    <pkg>            show package metadata
```

### Options

```
--pkgmgr apt|pacman   force a package manager (default: auto-detect)
```

Detection order (mirrors `termux-setup-package-manager`):
`TERMUX_APP_PACKAGE_MANAGER` env → `TERMUX_MAIN_PACKAGE_FORMAT` env → presence
of `pacman`/`apt` on `$PREFIX/bin` → fall back to `apt`.

`tvpkg` refuses to run as root, the same policy as the stock `pkg` script.

### Dry run

Set `TVPKG_DRY_RUN=1` to print the commands tvpkg would run without executing
anything.

## Build

```sh
make build     # builds ./tvpkg
make vet       # gofmt + go vet
make test      # go test ./...
make release   # cross-compiles all 4 Termux arches (needs NDK clang)
```

CI (`.github/workflows/release.yml`) builds `aarch64`, `arm` (GOARM=7),
`x86_64` and `i686` (GO386=sse2) binaries on every `v*` tag and attaches them
to a GitHub Release.

## License

MIT