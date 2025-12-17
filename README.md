# Spuderman 🕷️🥔

Spuderman spiders SMB shares (and local paths) to find sensitive files based on filenames, extensions, and content.

## Installation

### Binary (Recommended)
Download the latest binary for your operating system from the [Releases](https://github.com/your-username/spuderman/releases) page.
- **Windows**: `spuderman-windows-amd64.exe`
- **Linux**: `spuderman-linux-amd64` or `spuderman-linux-arm64`

### Build from Source
If you prefer to build from source, ensure you have [Go](https://go.dev/dl/) installed (1.21+ recommended).

1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/spuderman.git
   cd spuderman
   ```

2. Build the binary (creates `spuderman` or `spuderman.exe` in current directory):
   ```bash
   go build .
   ```

   **Cross-Compilation Examples:**
   - Windows: `$env:GOOS='windows'; $env:GOARCH='amd64'; go build .` (PowerShell)
   - Linux: `GOOS=linux GOARCH=amd64 go build .` (Bash)

## Usage

Basic usage:
```bash
./spuderman [flags] [targets]
```

Spuderman can spider specific hosts, CIDR ranges, or targets from a file.

**Targets supported:**
- Single IP/Hostname: `192.168.1.1`
- CIDR Range: `192.168.1.0/24`
- File of Targets: `targets.txt` (One target per line)

### Examples

**Spider a single host for passwords in files:**
```bash
./spuderman -u 'user' -p 'password' -d 'domain' -c 'password' 192.168.1.10
```

**Spider multiple hosts for specific extensions:**
```bash
./spuderman -u 'user' -p 'password' -d 'domain' -e 'kdbx,docx,xlsx' 192.168.1.10 192.168.1.11
```

**Use current user credentials (SSO/Kerberos):**
```bash
./spuderman -k 192.168.1.10
```
*(Note: Requires valid TGT or correct environment setup)*

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--username` | `-u` | Username for authentication | |
| `--password` | `-p` | Password for authentication | |
| `--domain` | `-d` | Domain for authentication | |
| `--hash` | `-H` | NTLM hash | |
| `--threads` | `-t` | Threads per host | 5 |
| `--parallel` | `-P` | Max concurrent hosts | 5 |
| `--maxdepth` | `-m` | Maximum spider depth | 10 |
| `--extensions` | `-e` | Filter by file extensions | |
| `--filenames` | `-f` | Filter by filename regex | |
| `--content` | `-c` | Search content regex | |
| `--loot-dir` | `-l` | Directory to save downloads | `.spuderman/loot` |
| `--no-download` | `-n` | Don't download files | `false` |
| `--analyze` | `-A` | Analyze mode (verbose, no download) | `false` |
| `--verbose` | `-v` | Show debug output | `false` |

## License
[MIT](LICENSE)
