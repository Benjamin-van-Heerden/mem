# Developing mem

This is the source repository for mem. It is not yet initialized as a mem-managed project.

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
- Copy the general guidelines below into generated project instructions as well as following them in this source repository.
- Go-specific test layout: preserve the existing colocated `_test.go` files for package tests; use `tests/` for standalone cross-package scenarios when appropriate. This is the source repository exception to the general default below.
- Intentional CLI stdout is product behavior; the restriction on diagnostic print statements does not prohibit it.
- Preserve user changes, stop immediately when asked to stop, and omit agent attribution from commit messages.

---

# General Principles

## Key Principle

Remember: "Whenever I'm about to do something, I think, 'Would an idiot do that?' And if they would, I do not do that thing." - Dwight Schrute

## Communication Style

- Be conversational but professional
- Think through considerations and requirements before writing code
- Planning first, then execution - we discuss the problem before implementing
- Don't be afraid to ask for help or input
- If you are unsure or need to guess about something, please ask

## Code Quality Standards

- Code should be self-explanatory - NEVER add comments unless absolutely necessary
- Avoid print statements apart from ad-hoc testing, when necessary defer to formal logging
- Follow established patterns and conventions in the codebase
- Prioritize clarity and maintainability over cleverness

## Performance Considerations

- Chunked processing for batch operations when applicable
- Database query optimization with proper indexing
- Memory management for large batch processing

## Modular Design

- Separate concerns into focused modules
- Robust error handling wherever applicable

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

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:

- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

### 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
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

---

# Testing Philosophy

Testing is critical to maintaining software quality, but not all tests are created equal. Focus on testing meaningful functionality that could actually break and impact the application.

### Test Structure

- Tests live in `tests/` directory with mirrored source structure
- Focus on meaningful functionality that could realistically break
- Avoid "idiot tests" that test framework behavior or trivial logic

### What to Test

- **Business logic**: Complex algorithms, validation rules, data transformations
- **API endpoints**: Request/response handling, authentication, error cases
- **Database operations**: Query correctness, constraint validation, data integrity
- **Integration points**: External API calls, file processing, third-party service communication

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

If the answer is no, delete the test and focus on more valuable testing efforts.

DELETE tests that don't follow these principles. NO 'IDIOT TESTS'!

NEVER run a full test suite unless specifically asked to. focus on specific tests related to the feature/functionality you are working on.
