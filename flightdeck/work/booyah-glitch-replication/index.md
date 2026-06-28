# Index — booyah-glitch-replication

## State

Booyah Glitch full-project from-scratch replication is built: all 12 comps are generated from Go API code, dual-version AE gated, and tracked under `flightdeck/showcase/booyah-clone/`.

User real-machine review is complete for comps 1, 2, 3, and 8. Comps 4, 5, 6, 7, 9, 10, 11, and 12 remain at review-gate status until the user validates them on the target machine.

## Next

Resume with user real-machine review for comps 10, 11, and 12. If accepted, update `flightdeck/showcase/booyah-clone/INDEX.md`, then close Task 4.3 in `plan.md` and sync `flightdeck/cockpit.md`.

## Read now

- `plan.md` — detailed comp-by-comp progress, verification history, and remaining Task 4.3 closeout
- `flightdeck/showcase/booyah-clone/INDEX.md` — review-gate ledger and documented fidelity deltas

## Read if

- `design.md` — if the replication success criteria or "no original chunk copy" method is questioned
- `flightdeck/knowledge/showcase/showcase.md` — before changing showcase status or adding review artifacts
- `flightdeck/knowledge/workflow/delivery-contract.md` — before claiming the replication is complete or usable
- `flightdeck/knowledge/workflow/ae-automation-occlusion-crashstate.md` — if AE automation hits modal/crash-state/occlusion failures

## Progress

Done:
- 12-comp Go generator and AE verification harness exist under `flightdeck/showcase/booyah-clone/`.
- Dual-version AE acceptance passed for the built replication.
- Known deltas are documented instead of hidden: Noise2 gap, Curves arbitrary-data block, Exposure wiggle tradeoff, eased Position builder gap, and remaining fidelity polish.

Current:
- Waiting on user real-machine review for 10, 11, and 12, then final review-gate bookkeeping.

## Open questions

- Whether the documented fidelity deltas should remain as accepted limitations or become follow-up polish tasks.
- Whether comp 1 still needs an AE 2020 render-side comparison.
