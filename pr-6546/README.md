# PR #6546: CodeBuddy Code review evidence

Reviewed head: `c9e14d9661257a429530aa0213e27bc0c9ab6e7e`.
Skill #7081 head: `04c8c95b23d717ccd2bc314e4577a0596b92e840`.
Original detailed review: https://github.com/OrchestratorInc/agent-orchestrator/pull/6546#issuecomment-6369710692
Native setup: @tencent-ai/codebuddy-code 2.164.0.

These are the retained 2026-10-23 Linux VPS review runs, not a new qualification run.
The product source was unchanged. Tests used isolated AO data, private native installs and a real Electron app.
Temporary sessions/apps, native installs, app state and worktrees were cleaned up after the runs.
Only selected test evidence is published; private environment/configuration files are excluded.

## Interpretation of raw runner output

- Already-installed screenshots do not prove the AO Install button works; that action was not exercised.
- The backend freeform model API passed. A catalog-centric Models-tab FAIL is not by itself a backend model bug.
- The independently verified New Task issue is the missing usable direct model entry when the backend says `customModelEntry: direct`.
- Authentication remained unknown/inconclusive. This is not proof the user is logged out.
- Provider-authenticated lifecycle, role, prompt, usage and successful history/restore checks were BLOCKED/SKIP.
- The negative create -> immediate kill -> restore probe returned HTTP 500 before a native history ID existed.
  It does not qualify positive history restoration or process-kill recovery.
- Update is NOT_RUN. Raw screenshot names and runner verdicts are preserved.

The CodeBuddy Login runner FAIL is a selector observation; the final manual control check timed out. The detailed review does not assert a separate proven missing-setup-button bug.

## File handling

PNG bytes are unchanged. Published text has ANSI styling removed and private test-root prefixes normalized. Original and published SHA-256 values are recorded in the root manifest. Screenshot OCR and text credential-pattern scans found no matches; this is a limited scan, not an absolute guarantee.
