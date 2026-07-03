# Fixture Generators

`test_data/generators/` contains JSX scripts that author, mutate, or verify AEP
fixtures through After Effects.

## Naming

- `verify_*` scripts usually build a fixture and assert one shipped behavior.
- `re_*` scripts are reverse-engineering probes that produced fixture evidence.
- `probe_*`, `build_*`, and `gen_*` scripts are support tools for focused data
  extraction or template generation.
- Versioned names such as `ae2020`, `ae24`, or `v2_2` describe the fixture
  generation context. They do not prove the script is obsolete.

## What Belongs Here

- JSX that can regenerate a tracked fixture or validate a stable behavior.
- Reverse-engineering JSX when its output is cited by a fixture, knowledge note,
  or registry evidence.
- Generator scripts that still document how a binary fixture was produced.

## Cleanup Rules

- Before deleting a generator, search for references in fixtures, tests,
  knowledge, registry evidence, and scripts.
- If a generator is superseded but still explains a shipped fixture, keep it or
  add replacement provenance first.
- Do not keep scripts that only produced disposable `tmp/` output and have no
  durable reference.
