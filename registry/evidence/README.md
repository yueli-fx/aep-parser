# Registry Evidence

`registry/evidence/` holds durable evidence files that are referenced by the
registry or by completed knowledge notes.

## What Belongs Here

- Topic-scoped evidence directories with stable filenames.
- Small reports, summaries, or proof artifacts that support registry facts.
- Evidence that should be reviewable in Git and reusable by future audits.

## What Does Not Belong Here

- Raw local output from exploratory commands.
- Heavy binary corpora or render outputs.
- Evidence that has no registry, knowledge, or task reference.

If an artifact only explains a temporary investigation, keep it in `tmp/` until
the finding is promoted into registry or knowledge.
