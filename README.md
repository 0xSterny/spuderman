# Spuderman 🥔🕷️

```text
 ____  ____  _   _ ____  _____ ____  __  __    _    _   _ 
/ ___||  _ \| | | |  _ \| ____|  _ \|  \/  |  / \  | \ | |
\___ \| |_) | | | | | | |  _| | |_) | |\/| | / _ \ |  \| |
 ___) |  __/| |_| | |_| | |___|  _ <| |  | |/ ___ \| |\  |
|____/|_|    \___/|____/|_____|_| \_\_|  |_/_/   \_\_| \_|
```

<div align="center">
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
-   **Memory Safe**: Limits file read sizes to prevent OOM on large files.

## Installation

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
  spuderman [flags] <target>

Flags:
  -H, --host string       SMB Host IP/Hostname
  -S, --share string      SMB Share Name
  -u, --user string       SMB Username
  -p, --pass string       SMB Password
  -d, --domain string     SMB Domain
      --hash string       NTLM Hash (Format: LM:NT)

  -c, --content string    Regex for content search
  -f, --filename string   Regex for filename search
  -e, --ext string        Comma-separated extensions (e.g., txt,pdf,docx)
      --preset string     Secret presets (aws, azure, slack, google, keys, auth)

  -o, --output string     Output file for results (JSON)
      --resume string     Resume state file (save/load progress)
      --loot string       Directory to download matched files (default "loot")
      --no-download       Only report matches, do not download
      --no-exclude        Disable default exclusions (node_modules, .git, etc.)

  -t, --threads int       Number of concurrent threads (default 4)
  -v, --verbose           Enable verbose logging
      --debug             Enable debug logging
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
Scan an SMB share for PDF files:
```bash
spuderman -H 192.168.1.10 -S C$ -u admin -p password -e pdf
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
