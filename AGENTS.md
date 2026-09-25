<mem>
<!-- Managed by mem v0.2.0. Edits inside this block are replaced on onboard. -->

# Working With mem

## Getting Started

Run `mem onboard` at the start of every session, and whenever the user says something like "let's get to work" or "get onboarded". It syncs this checkout with the shared codebase, applies mem updates and prints the project context: the codebase structure doc, project docs, runnable output, active specs, open todos, recent work logs and release status. Read all of it and follow its agent instruction. Use `mem onboard --offline` when there is no network access.

Don't rerun onboard later in the session unless the user asks.

mem commands print instructions specific to the current state. Follow them. Use `mem <command> --help` for details.

## Staying in Sync

Everyone works on the same codebase. Work on the development branch. Commit each coherent, working change, typically one commit per task or fix, rather than every edit or a whole day's work. Push at the end of each session together with the work log, after completing a spec, and before a promotion. mem applies safe Git updates itself and flags drift with ⚠️. Tell the user about each flag and help resolve it promptly. Run `mem sync` to fetch and update mid-session, for example before starting new work.

## Specs, Tasks and Todos

Ordinary coding needs none of these. Use them when the user asks for planned work or wants something tracked.

**Specs** are larger pieces of planned work, broken into ordered **tasks**:

- `mem spec new "<title>"`: create a draft spec, then follow its instructions to write it up with the user.
- `mem task new "<title>" "<detailed description>" --spec <spec>`: add a task.
- `mem spec start <spec>`: begin implementation. This assigns the spec to you and publishes that.
- `mem task complete <task> "<what was done and how it was verified>"`: record a finished task, then continue with the next one without waiting for approval.
- `mem spec complete <spec>`: once every task is done and the Success Criteria hold in the code.
- `mem spec list`, `mem spec show <spec>`: see what exists and where it stands.

`--spec` can be omitted when exactly one active spec is assigned to you. Refer to specs, tasks and todos by slug or by title.

**Todos** are standalone matters that need attention:

- `mem todo new "<title>" "<description>"`: record one.
- `mem todo list`, `mem todo show <todo>`: see open todos.
- `mem todo claim <todo>`: claim it as soon as you start working on it.

## Work Logs

Work logs carry context from one session to the next: what was done, what failed and what comes next.

- `mem log new`: create a log for this session and fill it in as instructed.
- Write a log at the end of a session, or when the session is getting long. Ask the user first.
- `mem log list`, `mem log show <log>`: read earlier logs.

## Codebase Structure

`.mem/structure.md` is a living map of the codebase, and onboard includes it. Keep it current rather than regenerating it. When your changes alter what it describes, such as modules, entry points, data flow or commands, update the affected sections and commit them with the work. `mem structure` creates the doc or lists what changed since it was last updated.

## Memories

Memories are lasting project conventions, kept in the memories section at the end of this file. They are not for session progress.

- When the user asks you to remember something, run `mem memory set <name> "<convention>"`. Use a short lowercase name, and reuse it to update that convention.
- When you notice a useful convention, suggest it, and record it if the user agrees.
- `mem memory remove <name>`: retire a convention.

Use project memories rather than external memory tools.

## Releases

Nothing deploys from the development branch. Releases move the staging and production branches forward along the development history, and CI deploys them.

- When the user asks for a preview release: `mem promote staging`. Use `--to <commit>` to leave out unfinished work.
- When the user asks for a release: `mem promote production`, then follow its instructions for the release notes.
- A full deployment means both, in that order.

Staging and production only move through `mem promote`. Never commit on them or push to them directly.

## Working Guidelines

- Run commands from the repository root unless instructed otherwise.
- Preserve changes you did not make and incorporate them into your understanding.
- When the user asks you to stop, stop immediately and wait for instructions.
- Keep commit messages descriptive, without agent names or co-authorship attribution.

---

# General Principles

## Key Principle

Remember: "Whenever I'm about to do something, I think, 'Would an idiot do that?' And if they would, I do not do that thing." - Dwight Schrute

## Communication Style
- Be conversational but professional
- Think through considerations and requirements before writing code
- Planning first, then execution - we discuss the problem before implementing
- Surface uncertainty and tradeoffs clearly, following the Think Before Coding guidelines below.

## Code Quality Standards
- Prefer self-explanatory code. Use comments to explain non-obvious reasoning, constraints, or tradeoffs; avoid comments that merely narrate what the code does.
- Avoid print statements apart from ad-hoc testing, when necessary defer to formal logging
- Follow established patterns and conventions in the codebase
- Prioritize clarity and maintainability over cleverness

## Performance Considerations
- Chunked processing for batch operations when applicable
- Database query optimization with proper indexing
- Memory management for large batch processing

## File and Folder Structure

- Follow the project's established layout and framework conventions. For new areas, organize files and folders around cohesive domains or features.
- Give each source file a focused responsibility. Keep related code together and separate concerns that change for different reasons.
- Start with a shallow structure. Introduce subfolders when they clarify meaningful boundaries, not merely to categorize a few files.
- Keep helpers and types close to the domain that owns them. Move code into shared modules only when it serves multiple areas; avoid catch-all `utils`, `helpers`, or `common` files.
- Prefer descriptive names that make a file's purpose clear without opening it.
- Keep dependencies between modules explicit. Avoid circular imports and splits that require unrelated modules to coordinate through shared mutable state.
- Use robust error handling wherever applicable.

### File Size and Refactoring

If a source file grows past 500 lines, treat that as a strong signal to split or refactor. Large files often accumulate mixed concerns, duplicated helpers, and hard-to-reason-about state. When your changes push a file past this threshold, or substantially extend an already oversized file, identify a cohesive part to extract before adding more.

Prefer extracting by domain or responsibility: domain-specific types, cohesive helpers, or a focused submodule. Do not split files mechanically to meet the line count, compress code to hide its size, or introduce unnecessary indirection. A useful split should make each module easier to understand and change independently.

Exceptions should be exceedingly rare, such as generated code or externally maintained vendored components. Explain why an exception is justified; compilation and passing tests alone do not justify keeping a file whole.

Keep refactoring within the requested scope. If an existing oversized file needs a broader restructuring than the task warrants, flag it and propose a follow-up rather than silently expanding the task.

## Functional Approach
- Prefer functional and procedural programming patterns over heavy OOP
- OOP is only used when it provides clear benefits
- Minimal abstractions - prefer explicit over implicit, declarative over imperative

### Source Formatting

- Do not hard-wrap lines just to satisfy an arbitrary line length. This project assumes modern editors with line wrapping.
- Keep user-facing strings, command strings, markdown output fragments, and simple expressions on one line when that is clearer.
- Only split a line when it improves structure or readability, such as a genuinely complex expression, a long data literal, or nested call arguments. When applicable use multiline strings for this.
- Do not reflow existing prose or strings unless the requested change requires it.

# Behavioral Guidelines

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

### 1. Think Before Coding

**Make assumptions explicit. Surface material uncertainty and tradeoffs.**

Before implementing:
- Check the available project context before asking for clarification.
- Ask when ambiguity would materially affect scope, behavior, architecture, or the consequences of an action. Explain the uncertainty and pause the affected work until it is resolved; continue independent work where useful.
- For routine, low-risk implementation choices within the authorized scope, follow established conventions and use judgment. State assumptions that affect the result without requiring confirmation for every minor decision.
- If a simpler approach exists, say so. Push back when warranted.

### 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- Introduce abstractions only when they clarify responsibilities or remove meaningful duplication. A focused function or module can be worthwhile with one caller; do not add layers solely for hypothetical reuse.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

### 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

### 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" -> "Write tests for invalid inputs, then make them pass"
- "Fix the bug" -> "Write a test that reproduces it, then make it pass"
- "Refactor X" -> "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:

```text
1. [Step] -> verify: [check]
2. [Step] -> verify: [check]
3. [Step] -> verify: [check]
```

Strong success criteria let you loop independently. Weak criteria like "make it work" require constant clarification.

---------------------------------------------------------------

# Testing Philosophy

Testing is critical to maintaining software quality, but not all tests are created equal. Focus on testing meaningful functionality that could actually break and impact the application.

### Test Structure
- Follow the project's established test layout and framework conventions. When establishing a new layout, default to a `tests/` directory that mirrors the source structure.
- Focus on meaningful functionality that could realistically break
- Avoid "idiot tests" that test framework behavior or trivial logic

### What to Test
- **Business logic**: Complex algorithms, validation rules, data transformations
- **API endpoints**: Request/response handling, authentication, error cases
- **Database operations**: Query correctness, constraint validation, data integrity
- **Integration points**: External API calls, file processing, inter-service communication

### What NOT to Test
- Framework internals 
- Third-party library behavior 
- Trivial getters/setters or simple data transformations
- Implementation details that don't affect public behavior

**Test Quality Principles:**
1. **Clarity Over Quantity** - Fewer, well-focused tests are better than many redundant ones
2. **Test Behavior, Not Implementation** - Focus on what the code does, not how it does it
3. **Meaningful Assertions** - Each test should verify something that could realistically fail
4. **Isolated Tests** - Tests should not depend on each other or external state
5. **Descriptive Names** - Test names should clearly describe what they're validating

**When in Doubt, Ask:**
- "Does this test validate critical business logic or user-facing behavior?"
- "Could this functionality realistically break in the way it is being tested?"

Apply this review to tests introduced or changed in the current task. Revise or remove those tests when they do not verify meaningful behavior, while preserving coverage required by the task. If an unrelated existing test appears unhelpful, flag it rather than deleting it without an explicit request.

NEVER run a full test suite unless specifically asked to. focus on specific tests related to the feature/functionality you are working on.
</mem>

<memories>
# Project Memories
</memories>

# Developing mem

This is the source repository for mem, and mem itself manages it: the installed `mem` on PATH runs the workflow here (onboard, specs, logs, hooks), while the source under development runs as `go run ./cmd/mem` or `dist/mem-dev`.

- Run development commands through `go run ./cmd/mem` or a build at `dist/mem-dev`.
- `mem` on PATH is the globally installed runtime. It does not necessarily execute the source being edited. There are no project-local runtime launchers in new projects.
- Test initialization, bootstrap, and upgrades in disposable Git repositories. Do not initialize or update this source checkout as a side effect of testing.
- Runtime assets are authored in `internal/agentsmd/instructions.md`. Never edit installed copies to implement source changes.
- `old/` is legacy reference material. Do not execute its onboarding or update commands.
- Design for solo developers and small teams in frequent contact. Drive convergence toward one shared codebase: surface branch drift early and guide timely synchronization while preserving uncommitted work. Use ordinary Git for sharing and conflicts; do not introduce a separate coordination or version-control system.
- Expose specs, tasks and todos directly through their own commands, accepting IDs or unambiguous names. Do not introduce a public umbrella work primitive.
- Keep stdout instructions grounded in actual state. Do not add generic workflow footers or print commands that are not implemented.
- Preserve the coding harness's stdout style: clear section headings and separators, readable context, and explicit, assertive agent instructions with concrete commands where warranted. Use `old/coding/.agent_core/harness/src/commands/onboard/formatting.py` and `content.py` as references. Carry forward the interaction style while adapting instructions to the current workflow and the user's existing authorization.
- Run focused Go tests for changes, and `go vet` for affected packages. End-to-end tests must not publish GitHub state or require credentials.
- The documents under `docs/` distinguish planned behavior from what is implemented. Preserve that distinction in output and documentation.
- Do not ship compatibility layers or historical snapshots for unreleased mem prototypes. Keep temporary development references outside the source tree. The explicitly supported Python-harness importer is separate from prototype compatibility.

## Persistent agent context

- Generated instructions at `internal/agentsmd/instructions.md` contain only what an agent needs routinely: working conventions and essential workflow entry points. They are not a CLI manual. Keep installation, initialization, migration, configuration, uncommon flags and operational preconditions in CLI help, documentation and state-specific stdout. Update the appropriate surface with each command change; do not add every command to persistent context.
- Treat compaction and session resumption as normal: an agent with AGENTS.md and repository state must retain the conventions and essential workflow entry points needed to continue. Detailed command usage is discovered through help and state-specific stdout. Dynamic progress needs durable work records once that feature exists; project memories are conventions, not task state.
- The general guidelines in the managed mem block are authored in `internal/agentsmd/instructions.md`. Follow them in this source repository too; change them there, never in this file.
- Go-specific test layout: preserve the existing colocated `_test.go` files for package tests; use `tests/` for standalone cross-package scenarios when appropriate. This is the source repository exception to the general test layout default.
- Intentional CLI stdout is product behavior; the restriction on diagnostic print statements does not prohibit it.
- Preserve user changes, stop immediately when asked to stop, and omit agent attribution from commit messages.
