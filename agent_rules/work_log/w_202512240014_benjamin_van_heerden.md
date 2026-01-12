# Work Log - AI Chat Streaming Fixes and Storage Analysis

## Overarching Goals
Continue refining the AI chat interface, specifically addressing UI issues with the streaming cursor and investigating the ballooning storage problem in the `ai_sessions` table caused by Agno's run storage architecture.

## What Was Accomplished

### Streaming Cursor Fix for Standalone Agents
Fixed a bug where the blinking cursor (`::after` pseudo-element) would persist after standalone agent runs completed.

**Root Cause:** The `RunCompleted` event handler only marked delegations as done, but didn't set `assistantMsg.isStreaming = false` for standalone agents.

**Fix in `useAiChat.js`:**
```javascript
case "RunCompleted":
case "RunResponseCompleted": {
  if (data.agent_id && activeDelegations[data.agent_id]) {
    activeDelegations[data.agent_id].isStreaming = false;
  } else {
    assistantMsg.isStreaming = false;  // Added for standalone agents
  }
  break;
}
```

### Removed Blinking Cursor Pseudo-Element
Removed the `::after` pseudo-element from `Content.vue` entirely as content streams in too quickly for it to be useful.

### Agno Storage Analysis
Analyzed a sample run object from the `ai_sessions` table and identified significant data duplication in Agno's storage architecture:

1. **Messages duplicated across runs** - Tool calls appear in both the `tools` array AND the `messages` array
2. **System prompts repeated** - Full system prompts stored per child run
3. **Tool data tripled** - Each tool call exists in: `tools` array, assistant message `tool_calls`, and tool response message
4. **Metrics duplication** - Token counts appear in session_metrics, run metrics, and message metrics

**Conclusion:** The Agno storage options (`store_member_responses`, `compress_tool_results`, etc.) have minimal effect. The core architecture duplicates data. User has decided to fork Agno to fix this.

## Key Files Affected

| File | Changes |
|------|---------|
| `frontend/composables/useAiChat.js` | Added `else` branch to set `assistantMsg.isStreaming = false` on `RunCompleted` for standalone agents |
| `frontend/components/ai/chat/messages/Content.vue` | Removed `.is-streaming ::after` pseudo-element and `@keyframes blink` animation |

## What Comes Next

1. **Database Indexes** - Still need to create composite indexes on `ai_sessions` to fix the "Out of sort memory" error:
   ```sql
   CREATE INDEX idx_sessions_team_list ON ai_sessions(user_id, session_type, team_id, updated_at DESC);
   CREATE INDEX idx_sessions_agent_list ON ai_sessions(user_id, session_type, agent_id, updated_at DESC);
   ```

2. **Agno Fork** - User plans to fork Agno repository to fix the storage architecture and eliminate data duplication in the `runs` column.
