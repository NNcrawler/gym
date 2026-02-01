# Agents.md

This document defines the configuration files and data contract used by the `gym` CLI.

For installation, usage, and command behavior, refer to `README.md`.

## Tech Stack

- Language: Go
- CLI framework: Cobra
- Config format: YAML
- YAML parser: `gopkg.in/yaml.v3`
- File operations: Standard library only
- Target platforms: macOS and Linux

---

## Global Configuration

The global config file is located at:

~/.gym.yaml

Format:

```yaml
skillRepository: /absolute/path/to/central/skills
````

`skillRepository` points to the local directory containing all centrally stored skills.
Each skill is a directory inside this repository.


---

## Project Configuration

Each project managed by `gym` contains a `.skills.yaml` file at its root.

This file declares:

* Which agents are used in the project
* Which skills are installed
* Optional per-agent custom installation paths

---

## `.skills.yaml` Format

```yaml
agents:
  - <agent-name>
  - ...

skillMap:
  <skill-name>:
    <agent-name>: <optional custom relative path>
```

### Example

```yaml
agents:
  - kilo-code
  - codex

skillMap:
  go-app-configuration:
    kilo-code: .kilocode/custom-skills/go-app-configuration
```

If an agent entry is missing under a skill, the CLI installs the skill into that agent’s default directory.

---

## Default Agent Directories

Default installation directories for each supported agent are maintained internally in the CLI codebase.

Examples:

* kilo-code → `.kilocode/skills/`
* codex → `.codex/skills/`

---

## Sync Model

The configuration assumes a one-way sync model:

central skill repository → project

Local modifications to installed skills are overwritten during synchronization.

---
