---
title: Promote staging wording when staging is not deployed
status: open
created_at: "2026-09-30T14:39:19+02:00"
---

Found in the nextjs-web end-to-end run (2026-09-30): mem promote staging tells the agent 'CI will deploy it. Once the preview checks out…', but the nextjs-web template deliberately does not deploy the staging branch (vercel.json git.deploymentEnabled test:false), so there is no preview to check and the agent was left unsure. Options: a [release] setting (e.g. staging_deploys = false) that changes the wording to 'staging is not deployed; promote production when ready', or wording that covers both cases. Keep stdout grounded in actual state.
