# split_roundtrip_baseline

Byte-identical round-trip safety net for V3 M8 physical split.

## Usage

Generate a fresh fingerprint file:

```
go run ./tmp_debug/split_roundtrip_baseline > tmp_debug/split_roundtrip_baseline/new.txt
```

Diff against the committed baseline (must be empty for round-trip invariant to hold):

```
git diff --no-index tmp_debug/split_roundtrip_baseline/baseline.txt tmp_debug/split_roundtrip_baseline/new.txt
```

## Expiry policy

`baseline.txt` is tied to the fixture set in `test_data/`. If fixtures are added, removed, or regenerated, regenerate `baseline.txt` with the command above and commit it with a message noting the fixture change, e.g.:

```
test(aep): regenerate round-trip baseline after fixture update
```

## Fixture git hash

`baseline.txt` was generated against fixtures present in the working tree at commit:

```
c45d7632e17676d21f737add9d5a1f9dbaaa0bbb
```
