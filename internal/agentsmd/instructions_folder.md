<mem>

# Working With mem

This project is a plain folder kept in sync by cloud storage. It is not a software project and it has no Git: there are no branches, commits or pull requests, and saving a file is all it takes. Do not set up Git here. Most of the work is documents, records and the thinking behind them, with a script here and there.

## Getting Started

Run `mem onboard` at the start of every session, and whenever the user says something like "let's get to work" or "get onboarded". It applies mem updates and prints the project context: the structure doc, project docs, runnable output, open todos and recent work logs. Read all of it and follow its agent instruction.

If `mem` is not installed (the command is not found), read `.mem/install.md` and follow all of it, including putting mem permanently on PATH, then run `mem onboard`. If it cannot be installed, tell the user before doing any other work.

Don't rerun onboard later in the session unless the user asks.

mem commands print instructions specific to the current state. Follow them. Use `mem <command> --help` for details. `.mem/README.md` explains how mem works here: the `.mem/` folder, its configuration, runnables, records and hooks. Read it when you need to know how a part of mem behaves.

## Skills and Automation

This folder is meant to grow into something that largely runs itself. Skills hold what gets reused here: formats, layouts, letterheads, procedures. Every task that will come up again should leave behind what handles it next time, so that what took a conversation today takes one instruction tomorrow. Building this up is a standing part of the work.

- **Look first.** Before starting a task, check `.claude/skills/`. Use the matching skill rather than working from memory or copying an older document.
- **Create skills without being asked.** When a task will clearly come up again, such as setting up a new client or adding an invoice template, make it a skill as part of doing it, then tell the user what you made. Do not ask first, and do not wait for it to happen twice. Work that really is a one-off stays a one-off.
- **Script what is mechanical.** Numbering, filling in, building, converting, filing and checking belong in a script, with the judgement left in the skill. Keep a script in the folder of the skill that runs it.
- **Keep skills true.** When a skill turns out wrong or incomplete, or the user corrects how something is done, change the skill in the same session as well as the work in front of you.

A skill is a folder under `.claude/skills/` holding a `SKILL.md`, whose description says when to use it: that is how a later session finds it. Keep real files there, not symbolic links.

Finished documents stand on their own. Take what a skill provides into the document itself rather than linking to a shared file, so that changing a skill never changes past work.

## Project Knowledge

- `.mem/structure.md` is a living map of this folder, and onboard includes it. When your work changes how the folder is organised, update the affected sections. `mem structure` creates it when it is missing.
- `.mem/docs/` holds Markdown reference documents, and onboard includes them in full. Keep them to what every session needs; anything needed only sometimes belongs in a skill or in the folder itself.
- Runnables are commands listed under `[runnables]` in `.mem/config.toml`, such as `python3 .mem/runnables/unbilled.py`, whose output joins the onboard context. Their scripts live in `.mem/runnables/`, where a file runs only when an entry names it. Use them for state worth seeing at the start of every session.

## Todos

Todos are matters that need attention, including anything waiting on someone else.

- `mem todo new "<title>" "<description>" --priority <critical|high|medium|low>`: record one. Choose the priority yourself unless the user names one: critical for what blocks work or leaves something broken, and onboard then asks for it to be taken up first; high for what should come next; medium (the default) for ordinary work; low for what can wait. `mem todo priority <todo> <priority>` changes it.
- `mem todo list`, `mem todo show <todo>`: see open todos.
- `mem todo delete <todo>`: once it is done or no longer relevant.

## Work Logs

A work log records a stretch of work: what was done, decided and tried, as fact. It is not updated later, so it never holds open work: anything still to be done, including blockers and decisions waiting on the user, is a todo.

- Write one when mem's work-log nudge says so and when the user wraps up. A log is a checkpoint, never a stop: write it at a natural stopping point, then carry on with the work. Run `mem log new` and fill the log in as instructed.
- Before you write a log, look back over the work: if any of it will come up again and no skill or script covers it yet, create that first.
- `mem log list`, `mem log show <log>`: read earlier logs.

## Memories

Memories are lasting project conventions, kept in the memories section of this file. They are not for session progress.

- When the user asks you to remember something, run `mem memory set <name> "<convention>"`. Use a short lowercase name, and reuse it to update that convention.
- When you notice a useful convention, suggest it, and record it if the user agrees.
- `mem memory remove <name>`: retire a convention.

Use project memories rather than external memory tools.

## Working Guidelines

- Run commands from the folder root unless instructed otherwise.
- Preserve changes you did not make and incorporate them into your understanding.
- When the user asks you to stop, stop immediately and wait for instructions.
- Keep moving on routine work. Ask only when the answer would change the result.

---

# General Principles

## Key Principle

Remember: "Whenever I'm about to do something, I think, 'Would an idiot do that?' And if they would, I do not do that thing." - Dwight Schrute

## Communication Style
- Be conversational but professional
- Think through what is being asked before producing anything
- For anything substantial, discuss the approach first, then do the work
- Surface uncertainty and tradeoffs clearly, and separate what the evidence shows from what you are assuming

## Documents, Reports and Visualizations
- Do not create hosted artifacts, canvases or similar pages outside this folder (such as Claude Artifacts), even when a tool or skill suggests publishing one. They are hard to edit, hard to fill with real data and live outside the project.
- Write documents, plans and reports as files in this folder, following its layout.
- Answers that only matter to the conversation stay in chat.

## Scripts

Scripts here are small tools, not products. Keep each one short and focused, and run it to check that it works.

# Behavioral Guidelines

**Tradeoff:** These guidelines bias toward care over speed. For trivial tasks, use judgment.

### 1. Think Before Acting

**Make assumptions explicit. Surface material uncertainty and tradeoffs.**

- Check the available project context before asking for clarification.
- Ask when ambiguity would materially affect the scope, the content or the consequences of an action. Explain the uncertainty, and continue with independent work where useful.
- For routine, low-risk choices within what was asked, follow established conventions and use judgment. State assumptions that affect the result without asking about every minor decision.
- If a simpler approach exists, say so. Push back when warranted.

### 2. Simplicity First

**The least that solves the problem. Nothing speculative.**

- Nothing beyond what was asked.
- No structure, process or options that were not requested.
- Skills and scripts for work that will come up again are the exception: they are expected unasked (see Skills and Automation). Build them for what the work has actually needed, and no further.
- Ask yourself: "Would a sensible colleague say this is overcomplicated?" If yes, simplify.

### 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

- Don't "improve" adjacent content, wording or formatting.
- Match the existing style, even if you'd do it differently.
- If you notice something unrelated that looks wrong, mention it rather than changing it.

The test: skills and scripts aside, every change should trace directly to the user's request.

### 4. Goal-Driven Execution

**Define what done looks like. Check it before you say so.**

- Turn a task into something you can verify: a document that builds and reads correctly, figures that reconcile to their source, a script that ran.
- Report what you verified and what you could not, plainly.
</mem>
