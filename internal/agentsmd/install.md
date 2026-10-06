# Installing mem

<!-- Managed by mem. Replaced on onboard; change it in the mem source, not here. -->

This project uses mem, a command-line tool that builds context for coding agents and keeps the project's work records in this repository. AGENTS.md says to run `mem onboard` at the start of every session. When `mem` is not installed, follow all four steps below, then run `mem onboard`.

**mem must end up permanently on PATH.** Every session starts with `mem onboard`, and agents run mem commands throughout the work. A mem that is only on PATH for the current session fails the same way in the next one. Step 2 is not optional.

- Install mem for the current user, without sudo or administrator rights.
- Step 2 changes the user's shell profile or PATH setting. Tell the user what you are adding and why, and make the change once they agree. If they decline, tell them that every new session will need this guide again.
- Once installed, mem updates itself at each `mem onboard`. Nothing else needs maintaining.
- Cloud and sandboxed agents whose environment is rebuilt for each session install mem the same way each time. Suggest to the user that they add the step 1 commands to the environment's setup script, so mem is there from the start.

## 1. Install

### macOS and Linux

With Go installed (`go version`), install it and note the directory it went to:

```sh
go install github.com/Benjamin-van-Heerden/mem/cmd/mem@latest
dir="$(go env GOPATH)/bin"
```

Without Go, download the release binary for this machine, check it against the release checksums and put it in `~/.local/bin`:

```sh
tag=$(curl -fsSLI -o /dev/null -w '%{url_effective}' https://github.com/Benjamin-van-Heerden/mem/releases/latest) && tag=${tag##*/}
os=$(uname -s | tr '[:upper:]' '[:lower:]')
case $(uname -m) in x86_64|amd64) arch=amd64 ;; arm64|aarch64) arch=arm64 ;; esac
name=mem_${tag}_${os}_${arch}
dir="$HOME/.local/bin"
tmp=$(mktemp -d)
(
  cd "$tmp" &&
  curl -fsSLO "https://github.com/Benjamin-van-Heerden/mem/releases/download/$tag/$name" &&
  curl -fsSLO "https://github.com/Benjamin-van-Heerden/mem/releases/download/$tag/checksums.txt" &&
  grep " $name\$" checksums.txt > expected.txt &&
  if command -v sha256sum >/dev/null; then sha256sum -c expected.txt; else shasum -a 256 -c expected.txt; fi &&
  mkdir -p "$dir" && chmod +x "$name" && mv "$name" "$dir/mem"
)
rm -rf "$tmp"
```

### Windows (PowerShell)

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
```

With Go, `go install github.com/Benjamin-van-Heerden/mem/cmd/mem@latest` works too; then set `$dir = "$(go env GOPATH)\bin"`.

## 2. Put mem on PATH permanently

### macOS and Linux

In the same shell as step 1, so `$dir` is set. This adds `$dir` to the profile files the user's shell reads, including the non-interactive shells agents run commands in, unless it is there already:

```sh
case $(basename "${SHELL:-sh}") in
  zsh) profiles="$HOME/.zshenv" ;;
  bash) profiles="$HOME/.bashrc $HOME/.profile" ;;
  *) profiles="$HOME/.profile" ;;
esac
for f in $profiles; do
  grep -qsF "$dir" "$f" || printf '\nexport PATH="%s:$PATH"\n' "$dir" >> "$f"
done
export PATH="$dir:$PATH"
```

### Windows (PowerShell)

In the same PowerShell session as step 1, so `$dir` is set:

```powershell
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($userPath -split ';') -notcontains $dir) { [Environment]::SetEnvironmentVariable('Path', "$userPath;$dir", 'User') }
$env:Path += ";$dir"
```

## 3. Check that a new session finds mem

macOS and Linux, in a fresh shell that starts from a clean environment and reads only the profile:

```sh
env -i HOME="$HOME" TERM=dumb "${SHELL:-/bin/sh}" -ic 'command -v mem'
```

Windows:

```powershell
([Environment]::GetEnvironmentVariable('Path', 'User') -split ';') -contains $dir
```

If the check prints nothing or `False`, step 2 did not take effect; fix it before continuing. Some tools (an IDE, a desktop app) only pick up a new PATH after they are restarted; tell the user if theirs needs one.

## 4. Continue

Run `mem version`, then `mem onboard`, and continue with what it prints.
