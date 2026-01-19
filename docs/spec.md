This technical specification defines **`zbwrap`**, a stateful management layer for the ZBackup deduplicating backup tool, optimized for local Go-based implementation.

---

## 1. System Overview

`zbwrap` serves as an orchestration wrapper that manages ZBackup repository locations, automates naming conventions, and persists human-centric metadata that is natively absent in ZBackup.

### Core Requirements

* **Statefulness**: Centralized registry for repository discovery.
* **Observability**: High-level reporting of repository sizes and backup inventory.
* **Automation**: Machine-readable JSON output and standardized naming schemas.

---

## 2. Data Architecture

### 2.1 Global Registry (`registry.json`)

Located at `~/.config/zbwrap/registry.json`, this file tracks all managed repositories on the local machine.

| Field | Type | Description |
| --- | --- | --- |
| `repositories` | Map | Keyed by logical alias; maps to physical filesystem paths. |
| `encryption` | Object | Stores encryption type (`none`, `password-file`) and credential paths. |
| `last_updated` | Timestamp | ISO-8601 string of the last registry modification. |

### 2.2 Metadata Sidecars (`<filename>.zbk.meta`)

Every backup artifact created by `zbwrap` is accompanied by a sibling JSON file to provide context without decompressing the main archive.

* **`mime_type`**: Detected via the first 512 bytes of the stream (e.g., `application/x-tar`).
* **`description`**: Optional user-provided string for human audit.
* **`status`**: Boolean or string indicating if the backup process finished successfully.

---

## 3. Functional Specification

### 3.1 Naming Strategy

New backups are enforced with the following pattern to ensure chronological lexicographical sorting:
`YYYY-MM-DD_HHMM-<suffix>.zbk`.

### 3.2 Error Handling & Atomic Safety

To maintain data integrity during execution:

* **Process Monitoring**: `zbwrap` monitors the ZBackup sub-process exit code.
* **Cleanup**: If ZBackup fails (non-zero exit), `zbwrap` must delete the associated `.meta.json` file to prevent stale metadata.

### 3.3 Information Commands

All information-oriented commands must support two output modes:

1. **Human Mode (Default)**: Formatted ASCII tables for CLI readability.
2. **Machine Mode (`--json`)**: Serialized JSON objects for piping into tools like `jq` or automation scripts.

---

## 4. Implementation Details (Go/Cobra)

* **CLI Framework**: Built using the **Cobra** library to handle nested subcommands (`add`, `backup`, `list`, `info`, `sync`).
* **Process Execution**: Uses `os/exec` to pipe `stdin` to ZBackup while simultaneously sniffing MIME types.
* **I/O Logic**:
  * **New Backups**: Capture metadata during the streaming process.
  * **Retroactive Sync**: A `sync` command to identify untracked `.zbk` files and generate metadata via side-loading.
