---
type: 'issue'
title: 'Issue: Panic on nil config'
number: 123
state: 'open'
author: 'alice'
created_at: '2026-01-01T10:00:00Z'
updated_at: '2026-01-02T11:00:00Z'
url: 'https://github.com/octo/repo/issues/123'
labels:
  - 'bug'
  - 'help wanted'
---

# Issue: Panic on nil config

## Metadata
- type: issue
- number: 123
- state: open
- author: alice
- created_at: 2026-01-01T10:00:00Z
- updated_at: 2026-01-02T11:00:00Z
- url: https://github.com/octo/repo/issues/123
- labels: bug, help wanted

## AI Summary

### Summary
The thread discusses root cause and fix.

### Key Decisions
- Use nil guard before dereference.
- Backfill regression tests.

### Action Items
- Release v1.0.1.
- Update documentation.

## Original Description

App panics when config is nil.

![image](https://example.com/a.png)

## Timeline
- 2026-01-01T10:00:00Z | opened | alice | Issue opened
- 2026-01-01T10:30:00Z | labeled | bot | bug
- 2026-01-01T11:00:00Z | assigned | maintainer | assigned to maintainer

## Discussion Thread
- bob (2026-01-01T12:00:00Z): I can reproduce this.
  - alice (2026-01-01T12:30:00Z): Thanks, investigating.
- carol (2026-01-01T13:00:00Z): Fixed in #124?

## References
- Original URL: https://github.com/octo/repo/issues/123
