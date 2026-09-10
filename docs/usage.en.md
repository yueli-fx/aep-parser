# SDK and integration guide

[中文首页](../README.md) · [English README](../README.en.md) · [中文指南](usage.md) · [English guide](usage.en.md)

## Go SDK

Once the repository is public, add the dependency from your external Go module:

```sh
go get github.com/yueli-fx/aep-parser
```

Import `github.com/yueli-fx/aep-parser`; the package name is `aep`.

This complete inventory example can run in your own Go module with a project path argument:

```go
package main

import (
    "encoding/json"
    "log"
    "os"

    aep "github.com/yueli-fx/aep-parser"
)

func main() {
    if len(os.Args) != 2 {
        log.Fatal("usage: inspect project.aep")
    }
    doc, err := aep.Open(os.Args[1])
    if err != nil {
        log.Fatal(err)
    }
    if err := json.NewEncoder(os.Stdout).Encode(doc.Inspect()); err != nil {
        log.Fatal(err)
    }
}
```

| API | Purpose |
| --- | --- |
| `Open(path)` / `Parse(io.ReadSeeker)` | Read a project with default resource budgets |
| `OpenWithLimits` / `ParseWithLimits` / `DefaultLimits` | Bound input/chunk size, cumulative allocation, node count, and recursion depth; zero fields use defaults |
| `Document.Inspect()` | Return a detached lightweight inventory |
| `Document.ProfileJSON()` | Return normalized profile JSON |
| `Document.ProjectJSON()` | Return a versioned property snapshot; see [ProjectJSON v2](project-json-v2.md) |
| `Document.Write(io.Writer)` | Round-trip a parsed project while preserving unknown chunks |
| `Document.Export(ctx, request, writer)` | Apply bounded property edits on a private copy, reparse, and verify unchanged bytes |
| `Compile(ctx, recipeJSON, writer)` | Create a new project and report capabilities, downgrades, refusals, and reparsing |

A `Document` **is not safe for concurrent use**. Use separate instances for concurrent tasks.

Export targets must come from `property_records[*].write_target` in the current document's snapshot. A `property_ref` only associates snapshot records; it is not a write locator. Export v1 accepts numeric static properties with actual writer backing and no keyframes, applying each batch on an all-or-nothing basis. Successful verification reports `byte-exact-outside-claimed-ranges`, meaning RIFX data and unknown chunks outside the declared edits remain byte-identical.

For both `Compile` and `Export`, check the Go **error and report `Status`**: `rejected` can be returned with a nil error. Use `CompileStatusVerified` / `ExportStatusVerified` to identify success and review downgrades. Writer I/O failures may leave partial output; file workflows should write a temporary file and replace the destination only after successful completion and close.

Recipe generation is not arbitrary ProjectJSON-to-AEP reconstruction. Unknown data can be preserved with an existing document without being semantically understood or independently reproducible.

## HTTP service

```sh
go run ./cmd/aepserver -addr 127.0.0.1:8080
```

| Method and path | Purpose |
| --- | --- |
| `GET /health` | Health status |
| `GET /capabilities` | Service capabilities, including unavailable host operations |
| `POST /parse` | Upload AEP bytes and obtain parsed data |
| `POST /profile` | Upload AEP bytes and obtain a normalized profile |

Upload with curl (use `curl.exe` in Windows PowerShell):

```sh
curl -H "Content-Type: application/octet-stream" --data-binary @project.aep http://127.0.0.1:8080/profile
```

The optional `X-AEP-Path` header labels the source path. Defaults include a 256 MiB upload limit, 32 concurrent requests, and HTTP timeouts. Server-side path input is disabled by default. To enable local batch workflows, explicitly pass `-allow-path-input -allowed-path-roots <root1,root2>` and submit JSON `{"path":"..."}`. Path mode checks directory and symlink escapes.

HTTP responses use their own schema: `/parse` is not the root SDK's ProjectJSON v2 endpoint. The service does not render or perform AE readback. See [SECURITY.md](../SECURITY.md) for untrusted-input and deployment boundaries.

## Compatibility and limitations

- AEP uses the big-endian RIFF container **RIFX / Egg!**. Parsing and writing rely on verified chunk layouts.
- The compatibility baseline is **After Effects 2020 (17.0)**. Versioned writers and migration workflows cover AE2020–AE2025. Newer files depend on their specific fields and layouts; universal forward compatibility is not promised.
- Reading, preserving, writing, creating from scratch, and passing host verification are separate capabilities. Consult the [capability matrix](capabilities.md).
- The core does not execute expressions, render frames, or replace third-party plugins, fonts, or footage. Confirm visual behavior in the target AE environment.
- Migration rebuilds supported semantics. Unknown-chunk preservation guarantees do not automatically apply to migration or Recipe generation.
- Some tests require local fixtures, AE installations, or private corpora and may skip. Passing pure Go tests does not imply complete host acceptance.
