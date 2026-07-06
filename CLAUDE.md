<!-- coact:begin -->
# coact collaboration contract (Claude Code)

You share this repository with another agent, coordinated by coact.

Your file edits are gated automatically by a coact PreToolUse hook:
- If an edit is denied with "coact: <path> is locked by <agent>", another agent
  is working there. Do NOT try to force it — switch to different files or wait.
  Run "coact status" to see who holds what.
- Allowed edits automatically record your lock; you never run "coact lock" by hand.

Divide the work explicitly instead of overlapping:
- Run "coact board" to see tasks, "coact claim <id>" before starting one, and
  "coact done <id>" when finished.
- Run "coact status" for live participants and locks, "coact log" for the audit trail.

This session should have COACT_AGENT=claude set.
<!-- coact:end -->
