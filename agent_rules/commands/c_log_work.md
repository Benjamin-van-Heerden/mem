$tool$ Get the current date and time

$tool$ Run (`git config user.name` ? `git config user.name` : `whoami`) |> toLowerCase() |> replaceSpacesWithUnderscores() $into$ --user_name 

$tool$ Create a work log file `agent_rules/work_log/w_{YYYYMMDDHHmm}_{--user_name}.md`

$if (there is an associated spec file)$
  $tool$ Read `agent_rules/commands/c_create_spec.md` so that you know what structure spec files should have.
  $tool$ Read the spec file to fully understand the requirements and context.
$end if$

$composite action$
  $tools: [edit file]$

  ~~ Edit the file with the work done based on our current interaction
  Work log files take on a specific structure, defined below:

  ```md
  # Work Log - {short title}

  $if (there is an associated spec file)$
  ## Spec File: `{spec_file_relative_path}`
  $end if$

  ## Overarching Goals
  {
  Broad goals and what we were trying to achieve with this work in the context of our interaction so far
  }

  ## What Was Accomplished
  {
  Description of what was done
  Use appropriate subtitles to organize work done and things achieved
  Don't mention anything that is not relevant to actual changes made, e.g. deliberations or context building actions
  You can be technical here and use actual code snippets and examples.
  }

  ## Key Files Affected
  {
  list of files affected and changes made
  }

  $if (there were errors or barriers)$
  ## Errors and Barriers
  {
  Implementation errors and barriers encountered that have not been resolved yet.
  Mention approaches which were tried and failed so we can learn from them and avoid repeating mistakes.
  }
  $end if$

  ## What Comes Next
  {
  If there are next steps or logical progressions from where we were, mention/list them here

  If we were working on a spec, mention the spec file here and where/how we could continue working on it. Mention which parts of the spec were completed and which parts need further work.
  }
  ```

  In writing work_log files be concise, but thorough enough so that someone else can pick up where you left off. It is important that you include any spec files that are associated with the work done, as the spec files are much more detailed and provide a comprehensive understanding what needs to be done and what has been done.

  work_log files will always be the *last thing* we do in our interactions to inform the next interaction.
$end composite action$

$if (there is an associated spec file)$
  $tool$ Edit the spec file to reflect the current state of the project.
$end if$

$finally$ Await further instructions, do not procceed on your own from here
