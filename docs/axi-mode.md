# AXI Mode for ACLI: Jira-First Plan

Status: approved design, planning only

## Scope

ACLI will gain an opt-in agent-oriented mode for Jira. Confluence and Bitbucket are explicitly out of scope for the first implementation, but the shared interfaces should leave room for them later.

The existing human and JSON behavior remains unchanged unless the caller opts into AXI mode.

## Goal

Make Jira state easy for agents to discover, consume, and act on without requiring a separate tool tree or interactive prompts.

The primary invocation is:

```text
acli --axi jira issue list
```

An `acli-axi` executable alias can be added later as a convenience for agent setup, but `--axi` is the canonical interface because it reuses the existing ACLI command tree.

## Current repository seams

The repository already provides the main foundations:

- `cmd/acli/root.go` owns persistent flags and the root command.
- `cmd/acli/output.go` selects text or JSON output and defines `OutputResult` for mutations.
- `cmd/acli/commands.go` exposes the command tree as machine-readable JSON.
- `cmd/acli/config.go` manages profiles and currently contains interactive setup flows.
- `cmd/acli/jira_issue.go` and `cmd/acli/jira_search.go` contain the highest-value Jira read and write paths.
- `internal/jira/` contains the API client and Jira domain operations.

AXI mode should be implemented at the output and context boundaries so existing API clients remain reusable.

## Invocation and compatibility

```text
acli jira issue list          # existing behavior
acli --output json jira issue list  # existing JSON behavior
acli --axi jira issue list    # AXI behavior
```

Rules:

1. Legacy invocations keep their current output and command behavior.
2. `--axi` selects the agent-oriented contract.
3. `--output json` remains an explicit escape hatch and overrides AXI rendering when requested.
4. AXI mode is non-interactive. Commands that require stdin return a structured error with a repair command.
5. Unknown flags fail loudly rather than being ignored.

## Jira context resolution

When a Jira command needs a project, resolve values in this order:

1. Explicit command arguments or flags.
2. A nearest `.acli/context.json` file with no secrets.
3. The selected profile's existing Jira project default.
4. A structured `context_missing` error with a concrete fix.

Example local context:

```json
{
  "profile": "work",
  "jira_project": "ENG"
}
```

The context file is for selectors only. Credentials remain in the existing profile configuration.

## Home view

Running AXI mode without a leaf command should return a bounded Jira-oriented snapshot instead of help text:

```text
bin: ~/.local/bin/acli
description: Agent-native Jira CLI
profile: work
site: https://acme.atlassian.net
context:
  jira_project: ENG
health:
  jira: ok
next[3]:
  Run `acli --axi jira issue list`
  Run `acli --axi jira search --jql <jql>`
  Run `acli --axi jira project get <project-key>`
```

The home view should perform only bounded, relevant checks. It must not fan out across every project or download broad Jira datasets by default.

## Output contract

AXI mode uses compact TOON-like records by default. The exact encoder should be hidden behind an internal package so the format can evolve without changing command handlers.

### Collections

List and search responses should include:

- the normalized query when applicable;
- `count`, `shown`, and pagination state;
- a minimal default field set of roughly 3-4 fields per item;
- explicit zero-result states;
- bounded `help[]` suggestions for the next likely commands.

Example:

```text
query: project = ENG AND status = "Open"
count: 3
shown: 3
issues[3]{key,status,summary,updated}:
  ENG-42,Open,"Fix token refresh",2026-09-08
  ENG-37,Open,"Document incident flow",2026-09-07
  ENG-21,Open,"Add Bitbucket link",2026-09-05
help[2]:
  Run `acli --axi jira issue get <issue-key>`
  Run `acli --axi jira issue create --summary <text>`
```

Large descriptions, comments, changelogs, and other bodies are truncated by default with a size hint. `--full` is the deliberate expansion path. `--fields key,status,summary` allows callers to request a specific schema.

### Mutations

Mutation responses should make the resulting state explicit and reusable:

```text
status: ok
action: jira.issue.create
key: ENG-43
message: Created issue ENG-43
data:
  url: https://acme.atlassian.net/browse/ENG-43
help[2]:
  Run `acli --axi jira issue get ENG-43`
  Run `acli --axi jira issue transition ENG-43 --transition <name>`
```

Destructive Jira operations require an explicit confirmation flag and never ask an interactive question.

### Errors

AXI errors should be structured on stdout, with diagnostics reserved for stderr:

```text
error:
  code: context_missing
  message: No Jira project is configured
  retryable: false
  fix: Set jira_project in .acli/context.json or pass --project <key>
  exit: 1
```

Recommended exit behavior:

- `0`: success;
- `1`: API, authentication, validation, or context error;
- `2`: invalid command or unknown flag.

## Jira-first command coverage

The first implementation should prioritize the workflows agents are most likely to chain:

1. `jira issue list`
2. `jira issue get`
3. `jira issue create`
4. `jira issue update`
5. `jira issue transition`
6. `jira issue comment list`, `add`, and `get`
7. `jira search`
8. `jira project list` and `get`
9. `commands` and the AXI home view

Remaining Jira resources can initially use a documented fallback renderer, but new commands should not silently omit AXI support. Add coverage tests as each resource is migrated.

## Implementation shape

Add an `internal/axi/` package with focused responsibilities:

```text
internal/axi/mode.go       mode and flag resolution
internal/axi/context.go    explicit, local, and profile precedence
internal/axi/record.go     collection, mutation, and error envelopes
internal/axi/toon.go       compact encoding, escaping, and truncation
internal/axi/hints.go      Jira-aware next-command suggestions
```

Refactor `cmd/acli/output.go` around a renderer interface:

```text
legacy text -> current tables and field views
legacy json -> current JSON payloads
axi         -> compact records, errors, and help[]
```

Keep API calls in `internal/jira/`. Command handlers should provide typed results and metadata to the renderer rather than hand-building AXI strings in every command.

## Agent guidance

Update `skills/atlassian-cli/SKILL.md` after the first implementation so the documented guidance and the AXI home view describe the same Jira command contract. Keep `acli commands` as the complete existing JSON tree and add a compact AXI capability response for `acli --axi commands`.

## Rollout

1. Add mode detection, the renderer boundary, structured envelopes, and tests.
2. Add Jira context resolution and the bounded home view.
3. Migrate issue list, get, create, update, transition, comments, search, and project commands.
4. Add truncation, field selection, pagination metadata, contextual help, and mutation safety gates.
5. Add coverage checks, documentation, and a Jira-only prerelease.
6. Consider Confluence and Bitbucket adapters only after the Jira contract has been exercised in real agent workflows.

## Success criteria

- An agent can discover the active Jira profile and project with one bounded home-view call.
- An agent can list, inspect, mutate, and follow up on Jira issues without reading a full command tree between calls.
- Large Jira content is bounded unless explicitly expanded.
- Empty results and failures are unambiguous and machine-readable.
- AXI mode never waits for stdin.
- Existing ACLI users and JSON scripts see no behavior change unless they opt in.
