# tvpkg — a friendly package manager for Termux

`tvpkg` (also available as `tvp`) is a wrapper around Termux's package
manager, supporting both **APT** and **Pacman**, with an interactive TUI and
a fast command-line interface. It is part of the
[TermuxVoid](https://termuxvoid.github.io/) project.

Real help is the package that ships with Termux (`pkg`); tvpkg behaves like
it, but adds a fuzzy-search TUI and works with either backend.

## Install

Install from the TermuxVoid repository:

```sh
pkg install tvpkg
```

Packages and prebuilt binaries are published at
<https://termuxvoid.github.io/>.

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

It opens on the package-manager home screen with a `tvpkg` banner (adapts to
the terminal size); start typing to pop up a small centered search window
(fzf/telescope style) with live results. Operations run on a clean progress
screen (spinner, progress bar, elapsed time) instead of flooding the terminal
with apt/pacman output; the raw log is still one key away.

Keybinds in the TUI:

| Key           | Action                                  |
|---------------|-----------------------------------------|
| `type`        | open search window and filter live      |
| `↑/↓`         | move up / down through results          |
| `←/→`         | cycle action (Install / Remove / Info)  |
| `PgUp`,`PgDn` | page up / down (also `Home`, `End`)     |
| `tab`         | cycle action (Install / Remove / Info)  |
| `enter`       | run the selected action                 |
| `y` / `n`     | confirm / cancel (when prompt is open)  |
| `esc`         | close the search window / close result  |
| `l` / `o`     | toggle the raw tool log while busy      |
| `q`, `ctrl+c` | quit                                    |
| touchscreen   | tap = select, double-tap = run, scroll = navigate |

The action pills always reflect the selected package's real state and the
list refreshes automatically after every operation:

- an **installed** package shows a green `●` and a disabled `installed` tag
  where the Install button would be — Install is skipped while cycling and
  never runs,
- **Remove** is dimmed for packages that are not installed,
- in Pacman mode **Info** is only offered for installed packages.

The header shows the package manager in use plus live `installed / available`
counts, and the Result window stays open after an operation until `esc`.

Non-interactive commands:

```
tvpkg install <pkgs...>   (i)  apt install -y / pacman -S
tvpkg remove  <pkgs...>   (r)  apt remove -y  / pacman -R
tvpkg search  <query>     (s)  apt-cache search / pacman -Ss
tvpkg update              (u)  apt update        / pacman -Sy
tvpkg upgrade             (up) apt update && full-upgrade / pacman -Syu
tvpkg list                (l)  list available packages
tvpkg list-installed     (li)  list installed packages
tvpkg info    <pkg>            show package metadata
tvpkg fix                    diagnose + repair interrupted databases
tvpkg files   <pkgs...>   (f)  dpkg -L / pacman -Ql
tvpkg clean                   apt clean        / pacman -Scc
tvpkg autoclean               apt autoclean    / pacman -Sc
```

Commands mirror Termux's `pkg` script (`$PREFIX/bin/pkg`): same underlying
backend operations, same aliases (`upg`, `cl`, `ac`, `li`, `f`).

### Options

```
--pkgmgr apt|pacman   force a package manager (default: auto-detect)
-y, --yes             skip all confirmation prompts
--simulate            dry-run: print commands without executing
```

Detection order (mirrors `termux-setup-package-manager`):
`TERMUX_APP_PACKAGE_MANAGER` env → `TERMUX_MAIN_PACKAGE_FORMAT` env → presence
of `pacman`/`apt` on `$PREFIX/bin` → fall back to `apt`.

`tvpkg` refuses to run as root, the same policy as the stock `pkg` script.

### Confirmation prompts

Mutating CLI commands (`install`, `remove`, `upgrade`) ask for confirmation
before proceeding.  A preview of the action (and upgrade plan, if applicable)
is shown so you can review what will happen.  Reverse-dependency warnings
are displayed when removing a package.

- **`-y` / `--yes`**: skip all prompts (useful in scripts).
- **`--simulate`**: preview the commands that would be executed, without changing
  anything — combines well with `-y` to silently list what would happen.
- **`TVPKG_DRY_RUN=1`**: environment-variable equivalent of `--simulate`.

In the TUI, install/remove actions open a prominent centered dialog — `press y
to run · n / esc to cancel` — before starting.  During a live operation,
`ctrl+c` asks for confirmation rather than quitting immediately: answering
**n** keeps the operation running, while **y** quits (which may leave the
database in a broken state).

### Fix command

```
tvpkg fix
```

Runs a quick database audit (`dpkg --audit` or `pacman -Dk`) and prints a
diagnostic.  If issues are found, it offers to run repair commands
(`dpkg --configure -a --force-confdef --force-confold` or
`pacman --noconfirm -D --asdeps …`) after a confirmation prompt.

### Audit log

All mutating operations (`install`, `remove`, `upgrade`, `repair`) are
logged to `$PREFIX/var/log/tvpkg.log` with a timestamp, command, packages,
and exit status.

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