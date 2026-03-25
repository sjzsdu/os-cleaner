# OS Cleaner

A cross-platform system cache cleaner for macOS and Linux.

## Features

- **Scan**: Discover how much space different caches are using
- **Top**: Find the largest files and directories in any folder
- **Inspect**: See what's actually inside each cache category
- **Active**: Detect which packages your projects are using
- **Clean**: Safely remove cache files with recoverable deletion support
- **Time-based filtering**: Find stale caches that haven't been accessed in a while

## Installation

```bash
# Clone or download this repository
cd os-cleaner

# Run the installation script
chmod +x setup.sh
./setup.sh
```

Or manually:

```bash
go build -o os-cleaner .
mv os-cleaner ~/.local/bin/
```

## Usage

### Scan All Caches

```bash
os-cleaner scan
```

Shows all cache categories and their sizes:

```
Total Cleanable: 78.6 GB
  - Safe: 65.4 GB
  - Caution: 13.2 GB
```

### Scan Stale Caches

Find caches not accessed in the last 30 days:

```bash
os-cleaner scan --stale
```

### Scan by Age

Find caches older than a specific time:

```bash
os-cleaner scan --older-than 90d   # 90 days
os-cleaner scan --older-than 6m    # 6 months
```

### Find Largest Files

Scan any directory to find the largest files and directories:

```bash
os-cleaner top
```

Output:
```
Disk Usage for: /Users/you/projects
Total: 2.3 GB

  Name                                                          Size      Type
  ────────────────────────────────────────────────────────────  ──────────  ────────
   1. node_modules                                               890.5 MB  dir (12453 files) ⚡
   2. .git                                                       456.2 MB  dir (8921 files)
   3. build                                                      234.1 MB  dir (342 files)
   4. video-asset.mp4                                            1.2 GB    file ⚠️

  Summary:
    1 item(s) ≥ 1GB (red)
    1 item(s) ≥ 100MB (yellow)
```

Scan a specific directory:

```bash
os-cleaner top ~/Downloads
os-cleaner top /var/log
```

Limit results:

```bash
os-cleaner top -n 10     # Show top 10 largest items
```

JSON output:

```bash
os-cleaner top --json
```

Color coding:
- **Yellow** ⚡ — items ≥ 100MB
- **Red** ⚠️ — items ≥ 1GB

### Inspect Cache Contents

See what's inside a specific cache:

```bash
os-cleaner inspect npm-cache
os-cleaner inspect go-modules
os-cleaner inspect python-pyenv
os-cleaner inspect homebrew-cache
```

### Detect Active Packages

Find which packages your projects are using:

```bash
os-cleaner active
os-cleaner active --path ~/projects
```

### Clean Caches

Preview what will be deleted:

```bash
os-cleaner clean npm-cache --dry-run
```

Clean specific category:

```bash
os-cleaner clean npm-cache
os-cleaner clean go-modules
os-cleaner clean homebrew-cache
```

Clean all safe categories:

```bash
os-cleaner clean --safe
```

### Recoverable Deletion

Compress files before deletion for recovery:

```bash
os-cleaner clean npm-cache --recoverable
```

Compressed files are saved to `~/.os-cleaner/recover/`

## Output Formats

JSON output for scripting:

```bash
os-cleaner scan --json
os-cleaner list --json
```

## Categories

### Development Tools

| Category | Description | Safety |
|----------|-------------|--------|
| Xcode DerivedData | Build intermediates | Caution |
| Xcode Simulator | iOS Simulator data | Caution |
| Docker | Images, containers, build cache | Caution |
| VSCode Cache | VSCode cache | Safe |
| JetBrains Cache | IntelliJ, PyCharm, etc. | Safe |

### Languages

| Category | Description | Safety |
|----------|-------------|--------|
| npm Cache | npm packages | Safe |
| yarn Cache | yarn packages | Safe |
| Python pip | pip packages | Safe |
| Python pyenv | Python versions | Caution |
| Go Modules | Go modules | Safe |
| Cargo Registry | Rust crates | Safe |
| Maven Repository | Java dependencies | Safe |
| Gradle Cache | Gradle caches | Safe |

### Package Managers

| Category | Platform | Description |
|----------|----------|-------------|
| Homebrew | macOS | Downloaded bottles |
| apt Cache | Linux | Debian/Ubuntu packages |
| dnf Cache | Linux | Fedora/RHEL packages |
| pacman Cache | Linux | Arch Linux packages |

### Browsers

| Category | Platform | Description |
|----------|----------|-------------|
| Chrome | All | Chrome cache |
| Firefox | All | Firefox cache |
| Safari | macOS | Safari cache |

## Safety Levels

- **Safe**: Can be deleted without any risk
- **Caution**: Can be deleted but may require rebuilds

## License

MIT
