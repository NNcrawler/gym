# gym

`gym` is a local CLI tool for managing and synchronizing reusable agent skills across multiple projects.

It solves a simple problem:

You maintain a **central repository of skills** on your machine.  
Each project can **pull selected skills** into agent-specific directories.  
Synchronization is **one-way**: central repository → project.

No network. No servers. No registries. Just allowing you to reuse skill implementations consistently across projects.

---

## Features

- Centralized local skill repository
- Per-project skill registration
- Agent-specific skill placement
- One-way synchronization
- Overwrite-on-sync behavior for deterministic updates
- Simple YAML configuration

---

## Concepts

### Central Skill Repository

All skills live in a single directory on your machine.  
Each skill is a directory containing its implementation files.

Example:

```

~/skills/
    go-app-configuration/
    postgres-introspection/
    http-client-helper/

```

The path to this directory is stored in a global config file:

```
~/.gym.yaml
```

Example:

```yaml
skillRepository: /Users/machine/skills
```

---

### Project Configuration

Each project using `gym` contains a `.skills.yaml` file at its root.
This file declares:

* Which agents are used in the project
* Which skills are installed
* Optional per-agent custom placement paths

Example:

```yaml
agents:
  - kilo-code
  - codex

skillMap:
  go-app-configuration:
    kilo-code: .kilocode/custom-skills/go-app-configuration
```

If no custom path is specified for an agent, `gym` uses the agent’s default skill directory.

---

### Default Agent Directories

Each supported agent has a default directory inside the project where skills are installed.
These defaults are maintained in the CLI codebase.

Examples:

| Agent     | Default skill directory |
| --------- | ----------------------- |
| kilo-code | `.kilocode/skills/`     |
| codex     | `.codex/skills/`        |

---

## Installation

Build and install the CLI:

```
go install github.com/nncrawler/gym@latest
```

Ensure `~/go/bin` is in your PATH.

---

## Global Setup

Create the global config file:

```
~/.gym.yaml
```

Example:

```yaml
skillRepository: /Users/machine/skills
```

---

## Usage

### Initialize a project

```
gym init
```

* Prompts for agents used in the project
* If `~/.gym.yaml` does not exist, prompts for the skill repository and creates it
* Creates `.skills.yaml`

---

### List available skills

```
gym list
```

* Reads the central skill repository from `~/.gym.yaml`
* Lists skill directories available to add
* Only includes directories containing a `SKILL.md`/`skill.md` file

---

### Add a skill

```
gym add <skill-name>
```

* Locates the skill in the central repository
* Copies it into the project for each configured agent
* Registers the skill in `.skills.yaml`
* Overwrites existing copies if present

---

### Remove a skill

```
gym remove <skill-name>
```

* Removes the skill from each configured agent directory
* Unregisters the skill from `.skills.yaml`

---

### Sync all skills

```
gym sync
```

* Reads `.skills.yaml`
* Re-copies each registered skill from the central repository
* Overwrites project copies

---

## Behavior Notes

* Synchronization is one-way
* Local modifications in project skill directories are overwritten
* Skills in the central repository are agent-agnostic
* Agent-specific placement is handled by `gym`

---

## Non-Goals

* Remote repositories
* Two-way sync
* Conflict resolution
* Version pinning

---

## License

MIT
