# zbwrap

`zbwrap` is a stateful, human-centric management layer for the **ZBackup** deduplicating backup tool.

ZBackup is excellent at deduplication and compression but lacks built-in management for repository locations, human-readable naming conventions, and metadata. `zbwrap` fills this gap by acting as an orchestration wrapper that adds a registry-based workflow, automatic naming, and metadata sidecars.

## Key Features

- **Centralized Registry**: Manage multiple ZBackup repositories using logical aliases.
- **Human-Centric Names**: Automatic enforced naming schema (`YYYY-MM-DD_HHMM-<suffix>.zbk`) for chronological sorting.
- **Metadata Sidecars**: Every backup is accompanied by a `.meta` JSON file containing MIME types, user descriptions, and success status.
- **Deep Inspection**: A `sync` command that can retroactively generate missing metadata and "deeply" sniff MIME types by restoring and probing archive headers.
- **JSON Support**: Machine-readable output mode (`--json`) for automation.

## Prerequisites

- **Go**: Version 1.20 or higher.
- **ZBackup**: Must be installed and available in your `PATH` or configured in the registry.
- **file**: Used for robust MIME type detection.

## Installation

```bash
# Clone the repository
git clone https://github.com/romiras/zbwrap.git
cd zbwrap

# Build the binary
make build

# Optional: Move to your PATH
mv zbwrap /usr/local/bin/
```

## Quick Start

### 1. Register a Repository
Initialize or link an existing ZBackup repository:
```bash
zbwrap add my-backups /path/to/backup/repo
```

### 2. Configure ZBackup Path (Optional)
If `zbackup` is not in your standard PATH:
```bash
# Edit ~/.config/zbwrap/registry.json to set "zbackup_path"
```

### 3. Create a Backup
Pipe data directly to `zbwrap`:
```bash
tar -cf - /home/user/data | zbwrap backup my-backups --suffix monthly --description "January Full Backup"
```

### 4. List Backups
```bash
zbwrap info my-backups
```

### 5. Synchronize Metadata
If you have old backups created via raw `zbackup`:
```bash
zbwrap sync my-backups --deep
```

## Architecture

- **Registry**: Stored at `~/.config/zbwrap/registry.json`.
- **Sidecars**: Metadata is stored alongside backups as `<filename>.zbk.meta`.
- **Logic**: Built with a hexagonal (ports and adapters) architecture to separate core logic from the CLI and ZBackup execution.

## License

[MIT](LICENSE)
