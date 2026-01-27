# Agents.md

This document defines the configuration format and behavior for the `gym` CLI, which manages synchronization of agent skills from a central local repository into individual projects.

The CLI is distributed separately. The central skill repository is a directory on the local machine. Synchronization is strictly one-way:

central skill repository → project

Local modifications to skills inside a project are overwritten during sync.

---

## Concepts

### Central Skill Repository

A single directory on the local machine stores all skills.
Each skill is a directory containing the files implementing that skill.

The path to the central repository is stored in a global config file:

```
~/.gym.yaml
```

### Global Config (`~/.gym.yaml`)

```yaml
skillRepository: /Users/machine/skills
```

Skills in the central repository do not contain agent-specific subdirectories.
Agent-specific placement is handled by the CLI.

---

### Project Skill Configuration

Each project using skills contains a `.skills.yaml` file at its root.
This file declares:

* Which agents are used in the project
* Which skills are added
* Optional per-agent custom target directories

### Example `.skills.yaml`

```yaml
agents:
  - kilo-code
  - codex

skillMap:
  go-app-configuration:
    kilo-code: .kilocode/skills-code/go-app-configuration
```

If a per-agent path override is not provided, the CLI uses the default directory for that agent.

---

## Default Agent Skill Directories

Each supported agent has a default directory inside a project where skills are placed. Defaults are maintained in the CLI codebase.

Examples:

* `kilo-code` → `.kilocode/skills/`
* `codex` → `.codex/skills/`

---

## CLI

Executable name:

```
gym
```

---

## Commands

### `gym init`

Initializes a project for skill management.

Behavior:

* If `~/.gym.yaml` does not exist, prompts for the skill repository and creates it
* Prompts the user to select which agents are used
* Creates a `.skills.yaml` file with selecteds the selected agents
* Does not copy any skills

---

### `gym list`

Lists available skills from the central repository.

Behavior:

* Reads central repository path from `~/.gym.yaml`
* Lists directories in the central repository
* Only includes directories containing a `SKILL.md`/`skill.md` file

---

### `gym add <skill-name>`

Adds a skill from the central repository into the project.

Behavior:

* Reads central repository path from `~/.gym.yaml`
* Locates `<skill-name>` directory in the central repository
* Copies the skill into the project:

  * For each configured agent
  * Into either:

    * The agent default skill directory, or
    * A per-agent override path if specified
* Registers the skill in `.skills.yaml`
* Overwrites any existing project copy of the same skill

---

### `gym remove <skill-name>`

Removes a registered skill from the project.

Behavior:

* Reads `.skills.yaml`
* Removes the skill directory for each configured agent
* Unregisters the skill from `.skills.yaml`

---

### `gym sync`

Synchronizes all registered skills.

Behavior:

* Reads `.skills.yaml`
* For each listed skill:

  * Copies from the central repository
  * Places into agent-specific directories
  * Overwrites existing project copies

Local modifications in project skill directories are overwritten without conflict resolution.

---

## Data Model

### `.skills.yaml`

```yaml
agents:
  - <agent-name>
  - ...

skillMap:
  <skill-name>:
    <agent-name>: <optional custom path>
```

If an agent entry is missing under a skill, the default directory for that agent is used.

---

## Supported Agents

The CLI maintains an internal registry of supported agents and their default skill directories. Adding new agents requires updating the CLI codebase.

---

## Non-Goals

* Two-way synchronization
* Conflict resolution
* Skill version pinning
* Remote repositories

---

## Tech

Tech stack
- Golang
- Cobra for CLI https://github.com/spf13/cobra
