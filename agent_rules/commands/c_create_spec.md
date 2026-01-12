$if (deliberations allow for spec creation, i.e. enough context about the spec/feature has been discussed)
  $continue$
$else$
  $stop$ Ask clarifying questions as is deemed necessary
$end if$

$tool$ Get the current date and time

$tool$ Run (`git config user.name` ? `git config user.name` : `whoami`) |> toLowerCase() |> replaceSpacesWithUnderscores() $into$ --user_name 

$tool$ Create a spec file `agent_rules/spec/s_{YYYYMMDD}_{--user_name}__{feature_name}.md`

$composite action$
  $tools: [edit file]$

  ~~ Edit the spec file based on deliberations.
  Work log files take on a specific structure, defined below:

  ```md
  # {Title of the spec}

  `%% Status: {Draft | In Progress | Completed | Archived | Abandoned} %%`

  ## Description
  {detailed description of what is to be done}

  ## Tasks

  $for each task$

  ### Task: {task title}
  - [ ] goal 1
  - [ ] goal 2
  - ...

  #### Implementation Details
  {
  implementation details in overview, specific code examples should be included only if the deliberations allow/ask for them
  }

  #### Testing Outline
  {testing outline}

  > Relevant existing files: [{list}]
  > New files: [{list}]
  > Tests: [{list}]

  ## Completion Report and Documentation
  {
  report on what was completed - when the specific task is completed
  ;;
  default='To be completed on task finalization'
  }

  $end for each$

  # Final Review
  {
  review of what was done, how it was implemented, issues encountered, future improvements and potential for directions to move forward - this is a final report, to be written only after the entire spec has been completed
  ;;
  default='To be completed on spec finalization'
  }
  ```

  Specs should be detailed and comprehensive, covering all aspects of the feature or task being implemented. They should include a clear description of the problem being solved, the proposed solution, and a detailed plan for implementation.
$end composite action$

$finally$ Suggest reductions in complexity, ask for more clarification, provide feedback on the spec's structure and content. The creation phase of a spec is iterative, and may require multiple rounds of refinement and feedback.
