# Installing mem

<!-- Managed by mem. Replaced on onboard; change it in the mem source, not here. -->

This project uses mem, a command-line tool that builds context for coding agents and keeps the project's work records in this repository. AGENTS.md says to run `mem onboard` at the start of every session. When `mem` is not installed, install it with these steps, then run `mem onboard`.

- Install mem for the current user, without sudo or administrator rights.
- Ask the user before changing anything that persists outside this repository and the install directory, such as a shell profile or the user's PATH setting.
- Cloud and sandboxed agents install it in their own environment the same way, at the start of each session.
- Once installed, mem updates itself at each `mem onboard`. Nothing else needs maintaining.

## macOS and Linux

With Go installed (`go version`):

```sh
go install github.com/Benjamin-van-Heerden/mem/cmd/mem@latest
```

This puts `mem` in `$(go env GOPATH)/bin` (usually `~/go/bin`), which must be on PATH.

Without Go, download the release binary for this machine, check it against the release checksums and put it in `~/.local/bin`:

```sh
tag=$(curl -fsSLI -o /dev/null -w '%{url_effective}' https://github.com/Benjamin-van-Heerden/mem/releases/latest) && tag=${tag##*/}
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case $(uname -m) in x86_64|amd64) arch=amd64 ;; arm64|aarch64) arch=arm64 ;; esac
name=mem_${tag}_${os}_${arch}
tmp=$(mktemp -d)
(
  cd "$tmp" &&
  curl -fsSLO "https://github.com/Benjamin-van-Heerden/mem/releases/download/$tag/$name" &&
  curl -fsSLO "https://github.com/Benjamin-van-Heerden/mem/releases/download/$tag/checksums.txt" &&
  grep " $name\$" checksums.txt > expected.txt &&
  if command -v sha256sum >/dev/null; then sha256sum -c expected.txt; else shasum -a 256 -c expected.txt; fi &&
  mkdir -p ~/.local/bin && chmod +x "$name" && mv "$name" ~/.local/bin/mem
)
rm -rf "$tmp"
```

If `command -v mem` still finds nothing, `~/.local/bin` is not on PATH:

- For the current session: `export PATH="$HOME/.local/bin:$PATH"`.
- To make it permanent, ask the user, then add the same line to their shell profile: `~/.zshrc` for zsh, `~/.bashrc` for bash.

## Windows (PowerShell)

```powershell
$tag = (Invoke-RestMethod https://api.github.com/repos/Benjamin-van-Heerden/mem/releases/latest).tag_name
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
$name = "mem_${tag}_windows_$arch.exe"
$base = "https://github.com/Benjamin-van-Heerden/mem/releases/download/$tag"
$dir = "$env:LOCALAPPDATA\Programs\mem"
New-Item -ItemType Directory -Force $dir | Out-Null
Invoke-WebRequest "$base/$name" -OutFile "$dir\mem.exe"
Invoke-WebRequest "$base/checksums.txt" -OutFile "$dir\checksums.txt"
$expected = (Select-String -Path "$dir\checksums.txt" -Pattern " $name$").Line.Split(' ')[0]
Remove-Item "$dir\checksums.txt"
if ((Get-FileHash "$dir\mem.exe" -Algorithm SHA256).Hash -ne $expected) { Remove-Item "$dir\mem.exe"; throw "mem.exe does not match the release checksum" }
$env:Path += ";$dir"
```

The last line puts mem on PATH for the current session. To make it permanent, ask the user, then run:

```powershell
[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User') + ";$dir", 'User')
```

## Check

Run `mem version`, then `mem onboard`, and continue with what it prints.
