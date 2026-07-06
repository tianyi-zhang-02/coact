<!-- coact:begin -->
# coact collaboration contract (Codex)

You share this repository with another agent, coordinated by coact. Your edits
are NOT gated automatically, so you MUST follow this protocol yourself.

Before editing ANY file or directory:
  1. Run: coact lock <path>
  2. If it prints "denied", STOP — another agent holds it. Do not edit. Choose
     other work or wait, and re-check with: coact lock <path> --check
  3. Edit only after the lock is granted.
When you are done with a path, run: coact unlock <path>

Divide the work explicitly:
- Run "coact board" to see tasks, "coact claim <id>" before starting, and
  "coact done <id>" after.
- Run "coact status" for who holds what, and "coact log" for the audit trail.

This session should have COACT_AGENT=codex set so the id is implicit.
<!-- coact:end -->
