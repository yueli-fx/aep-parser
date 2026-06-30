# Technique Report Dashboard Spec

## Goal

Turn the self-hosted `report.html` from a long generated dump into a human review dashboard.
The report should help a user understand a corpus quickly, then inspect only a few relevant
projects or patterns. The machine-readable artifacts remain the source of truth.

## Core Principle

Humans do not read the whole corpus. The HTML must default to summary, ranking, and guided
drilldown:

- show the corpus conclusion first;
- surface the few projects and patterns worth inspecting;
- keep long-tail records searchable but collapsed;
- open project details only when selected;
- leave full raw facts in JSON/CSV artifacts.

## Non-Goals

- Do not remove `summary.json`, `corpus.jsonl`, CSV exports, recipe drafts, or other artifacts.
- Do not require a web server; the report should stay a static local HTML file.
- Do not make a marketing page or prose-heavy narrative report.
- Do not render every comp/layer/effect detail into the initial page body.

## Information Architecture

The report is a single-page dashboard with tabs:

1. `Overview`
   - Corpus counts: parsed projects, parse errors, comps, layers, effects, patterns, recipe drafts.
   - Top findings: the most important learned patterns and why they matter.
   - Risk summary: parse errors, unknowns, third-party effects, low-confidence projects.
   - Next actions: capability gaps, projects to inspect, recipe candidates.

2. `Patterns`
   - Ranked pattern/mechanism table.
   - Filters for category, confidence, frequency, and blocker status.
   - Each row expands to a short explanation, representative projects, and supporting signals.
   - Default display is top-ranked patterns only.

3. `Projects`
   - Searchable project list.
   - Default filter shows representative, anomalous, failed, or high-value projects first.
   - Clicking a project opens detail in a dedicated panel instead of navigating a long page.
   - Detail includes summary, reconstruction blueprint, key comps, signal layers, effects,
     shape/text mechanisms, blockers, and linked raw artifacts.

4. `Recipes`
   - Recipe drafts and smoke-test state.
   - Show pass/fail, blocker reason, and the project/pattern that produced the draft.
   - Keep full recipe JSON behind links or collapsed code blocks.

5. `Data`
   - Artifact index with descriptions: `summary.json`, `corpus.jsonl`, `digest.json`, CSVs,
     `reconstruction_blueprints.jsonl`, `recipe_drafts.jsonl`, `report.md`, and manifest.

## Data Model

Do not change the canonical machine artifacts for this UI pass. `report.html` should be a
renderer over existing report artifacts:

- `summary.json` for corpus-level counts;
- `digest.json` for ranked findings and top-level learning signals;
- `projects.csv`, `patterns.csv`, `mechanisms.csv`, and related CSVs for tables;
- `corpus.jsonl` and `reconstruction_blueprints.jsonl` for selected project detail;
- `recipe_drafts.jsonl` for recipe candidates.

If a required field is missing, render a clear empty state rather than failing report generation.

## Rendering Strategy

Generate one self-contained HTML file:

- inline CSS and a small plain JavaScript controller;
- keep the UI's required table/detail data in a compact embedded JSON payload so double-clicking
  `report.html` works without a web server;
- keep generated artifacts linked from the `Data` tab for full raw inspection;
- render only overview and top rows at initial load;
- render project detail lazily after the user selects a project;
- avoid repeating the same large data in multiple DOM sections.

The generator should continue to live in `internal/selfhost/technique_report.go` unless the
renderer grows enough to justify mechanical extraction later.

## UX Requirements

- First viewport must answer: "What did this corpus teach us?"
- A user should be able to find a project by filename in one search box.
- A user should be able to inspect a representative project without scrolling through unrelated
  projects.
- Empty/error states must be visible and concise.
- Long code/JSON snippets must be collapsed by default.
- Tables should favor scan columns: name, count, confidence, blocker, representative project,
  action.

## Verification

Update self-host report tests to assert:

- `report.html` contains the tab labels and project-detail UI hooks;
- required artifact links still exist;
- report verification still passes for minimal fixtures;
- generated HTML includes representative project data but does not render all detail rows eagerly.

Run:

```powershell
go test ./internal/selfhost ./cmd/aepselfhost -count=1
go test ./... -count=1
go vet ./...
go run ./cmd/aepselfhost technique-report -limit 3 -verify
```

`technique-report -limit 3 -verify` is enough for this UI pass; full self-host verification can run
after the dashboard renderer is stable.
