---
title: Docs, proposal trim and deferred cloud-agent todo
status: completed
created_at: "2026-10-26T12:58:18+02:00"
updated_at: "2026-10-26T13:05:13+02:00"
completed_at: "2026-10-26T13:05:13+02:00"
---

Update docs/design.md (Ground Rules, push retry, local merges, [git] base in the config example), docs/status.md (implemented), .mem/structure.md (converge and onboard changes), README if it describes sync or config. Trim docs/cloud_agents.md: mark the cloud-agent parts as deferred and the rest as implemented by this spec, keep the research. Record a todo 'Cloud agent support' pointing at docs/cloud_agents.md section 2 (environment setup, detection, refusing releases in cloud, MEM_USER, switch-to-dev/landing, loud missing-mem hooks).

## Completion Notes

docs/design.md: base in the config example, table rows for local merges and retried pushes, a Long-lived branches paragraph, Ground Rules and the static block under Updates. docs/status.md: the three implemented items. README: base in the config example. .mem/structure.md: converge (catchUp, pushAttempts, hasMerges), git (ErrRemoteAhead), project (base), the onboard data flow (onboard_base.go, baseLine, agentsSizeWarning). docs/cloud_agents.md committed with its status: sections 1, 3 and 4 implemented (static block instead of a rendered one, and why), section 2 deferred, decisions taken. Todo cloud_agent_support records the deferred work. go vet ./... passes.
