# Spuderman 🥔🕷️

<div align="center">
<pre>
 ____  ____  _   _ ____  _____ ____  __  __    _    _   _ 
/ ___||  _ \| | | |  _ \| ____|  _ \|  \/  |  / \  | \ | |
\___ \| |_) | | | | | | |  _| | |_) | |\/| | / _ \ |  \| |
 ___) |  __/| |_| | |_| | |___|  _ <| |  | |/ ___ \| |\  |
|____/|_|    \___/|____/|_____|_| \_\_|  |_/_/   \_\_| \_|
</pre>
  <b>The potato-powered file spider and secret scanner.</b>
</div>

---

Spuderman is a high-performance, memory-safe file spider and content search tool written in Go. It is designed to scan local directories and SMB shares for interesting files, secrets, and patterns.

## Features

-   **Fast & Concurrent**: Multi-threaded scanning and processing.
-   **Protocol Support**: Local Filesystem and SMB (v1/v2/v3).
-   **Content Extraction**:
    -   Text files
    -   PDF Documents (OCR-like text extraction)
    -   Office Documents (DOCX, XLSX, PPTX)
-   **Secrets Detection**:
    -   Built-in presets for AWS, Azure, Google, Slack, Private Keys, and more.
    -   Custom regex support.
-   **Resumable Scans**: Save state and resume interrupted scans (`--resume`).
-   **Async Downloads**: Downloads matched files in the background without blocking the scan.
-   **Output Formats**: Console (Human-readable) and JSON (`--output`).
-   **Live Progress Bar**: A progress bar stays pinned to the bottom of the terminal while log output scrolls above it.
-   **Silent Mode**: `--silent` suppresses everything except matches and downloads — the positive hits.
-   **Memory Safe**: Limits file read sizes to prevent OOM on large files.

## Installation

### Go Install
Install the latest version directly with Go (requires Go 1.21+):

```bash
go install github.com/0xSterny/spuderman@latest
```

This places the `spuderman` binary in `$(go env GOPATH)/bin`. Make sure that directory is on your `PATH`.

### Binary Release
Download the latest binary from the [Releases Page](https://github.com/0xSterny/spuderman/releases).

### Build from Source
Requirements: Go 1.21+

```bash
git clone https://github.com/0xSterny/spuderman
cd spuderman
go build .
```

## Usage

```txt
Usage:
  spuderman [targets] [flags]

Targets can be a single IP/Hostname, a CIDR range, a file of targets
(one per line), or a local directory.

Flags:
  -A, --analyze              Analyze mode: No download, Verbose output, Log to file
  -b, --blacklist strings    Comma-separated substrings to exclude from results (path match, case-insensitive)
      --ccache string        Kerberos CCache file path
  -c, --content strings      Search for file content using regex
  -x, --delimiter string     Delimiter between Host/Share/Path in flat loot filenames (default "+")
      --dirnames strings     Only search directories containing these strings
  -d, --domain string        Domain for authentication
  -e, --extensions strings   Only show filenames with these extensions
  -f, --filenames strings    Filter filenames using regex
  -H, --hash string          NTLM hash for authentication
      --krb5-conf string     Kerberos config file path (krb5.conf)
  -l, --loot-dir string      Loot directory (default ".spuderman/loot")
  -m, --maxdepth int         Maximum depth to spider (default 10)
  -n, --no-download          Don't download matching files
      --no-exclude           Disable default exclusions
      --no-pass              Do not use a password (force empty)
  -o, --output string        Output file for results (JSON)
  -P, --parallel int         Max concurrent hosts (default 5)
  -p, --password string      Password for authentication
      --preset strings       Load secret regex presets (e.g. aws, azure, slack, keys)
      --resume string        Resume state file (JSON)
      --sharenames strings   Only search shares with these names
      --silent               Only show matches and downloads (suppress all other console output and the progress bar)
  -S, --structured           Use structured loot directory (Host/Share/File)
  -t, --threads int          Concurrent threads (PER HOST) (default 5)
  -u, --username string      Username for authentication
  -v, --verbose              Show debugging messages
```

## Examples

### 1. Basic Local Scan
Scan a directory for files containing "password":
```bash
spuderman -c "password" /path/to/scan
```

### 2. Secrets Scanning (Presets)
Scan for cloud keys and private keys:
```bash
spuderman --preset aws,keys /path/to/source
```

### 3. SMB Share Scanning
Scan a specific SMB share for PDF files (the target is positional):
```bash
spuderman -u admin -p password --sharenames C$ -e pdf 192.168.1.10
```

### 4. Resume Scan
Run a scan and save state to `progress.json`. If interrupted, run the same command to resume:
```bash
spuderman --resume progress.json -c "confidential" 192.168.1.0/24
```

### 5. JSON Output
Save findings to a structured JSON file for processing:
```bash
spuderman -o results.json --no-download 10.0.0.5
```

### 6. Multiple Keywords (Filename + Content)
Search filenames AND file contents for any of several credential-related keywords. Both `-f` and `-c` accept comma-separated values:
```bash
spuderman -f "pwd,passw,admin,login,logon" -c "pwd,passw,admin,login,logon" 192.168.1.0/24
```

### 7. Blacklist
Skip any file whose path contains one of the given substrings (case-insensitive). Useful for trimming noisy hits like browser caches or sample data:
```bash
spuderman -c "password" -b "node_modules,sample,test_data" /path/to/scan
```

## Presets
Available presets for `--preset`:
-   `aws`: AWS Access Keys, Session Tokens
-   `azure`: Azure Storage Keys, SAS Tokens
-   `google`: GCP API Keys, OAuth
-   `slack`: Slack Webhooks, Tokens
-   `keys`: Private Keys (RSA, DSA, EC, OpenSSH)
-   `auth`: Generic API Keys, Bearer Tokens, Basic Auth

## License
MIT
