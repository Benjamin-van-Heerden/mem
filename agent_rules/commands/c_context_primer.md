$tool$ Run `tree -I "spec|work_log" agent_rules`

$tool$ Run `git log -n 5`

$tool$ Run `git add . && git status`

$if (there are changes to commit)$
  $tool$ Run `git diff` (check what has changed)

  $tool$ Run `git commit -m "{appropriate_commit_message}"` (interpolate the commit message - be descriptive and detailed)

  $tool$ Run `git push`
$endif$

$tool$ Run `git ls-files ':(exclude).agent_rules/*'` in the terminal

-------------------------
Project Specific Actions
-------------------------

$tool$ Read project specific files
  - `./pyproject.toml`
  - `./README.md`
  - `./docs/ai_chat_implementation_plan_v4.tex`
  - `./docs/agent_platform_expansion_implementation_plan.tex`
  - `./docs/adr/adr_20251215_agno_sessions_over_custom_persistence.md`
  - `./src/api/main.py`
  - `./src/task_runner/main.py`
  - `./agent_rules/docs/d_final_instructions_and_reminders.md`

-------------------------

$tool$ Get the current date and time

$tool$ List the contents in `./agent_rules/work_log`

$tool$ Read the *3* most recent files in the work log

$if (there are spec files to read)$
  $tool$ Read the mentioned spec files
$else$
  $tool$ Run `git diff HEAD^ HEAD` (to see what has changed in the last commit)
$endif$

$finally$ Await further instructions, do not procceed on your own from here
