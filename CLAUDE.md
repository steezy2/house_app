# CLAUDE.md

Project navigation notes for Claude Code. Human-facing docs live in
[README.md](README.md), [house_app_server/README.md](house_app_server/README.md),
[house_app_server/API_EXAMPLES.md](house_app_server/API_EXAMPLES.md), and
[IMAGE_PROCESSING.md](IMAGE_PROCESSING.md) — this file is a faster map for
finding things and a list of non-obvious traps, not a replacement for them.

## What this is

A personal/home photo-library backend: a single Go API (no web or mobile
frontend lives in this repo — interact via curl or the generated Swagger
UI). The primary intended client is a phone on the home network — e.g. a
"select photos, tap upload" flow hitting `POST /images/multiple` — not a
browser or desktop app. Clients upload images; a background worker
separately reorganizes `./uploads` into a dated, categorized folder tree.
It started life as a bookmark-manager server (`bookmark_server`) — see
[MIGRATION_SUMMARY.md](MIGRATION_SUMMARY.md) for that rewrite's history if
you find a stray reference to bookmarks/scraping anywhere.

## Where things live

- `house_app_server/`: the entire application (single Go module,
  `house-app`). Layered `clients` (Mongo) -> `models` (domain types + repo
  interface) -> `service` (upload/business logic) -> `controller` (HTTP
  routes + swag annotations). Composition root: `main.go`.
  - `processor/image_processor.go`: independent background loop, started
    from `main.go` alongside (not through) the HTTP server. Polls
    `UPLOAD_DIR`, extracts EXIF, categorizes, renames, and moves files into
    `STORAGE_DIR/YYYY/MM/category/`.
  - `docs/`: generated Swagger/OpenAPI output (`docs.go`, `swagger.json`,
    `swagger.yaml`). Regenerate with `make docs` after touching swag
    annotations in `controller/api.go` or `main.go` — don't hand-edit.
- Root-level `*.md` files are the human docs; `MIGRATION_SUMMARY.md` is a
  point-in-time record of the bookmark→house-app rewrite, not living
  documentation — don't "fix" it to match current behavior.

## Key conventions and known traps worth knowing before touching this code

- **Auth is a single shared secret, not per-user.** If `API_KEY` is set,
  `controller.APIKeyMiddleware` requires a matching `X-API-Key` header on
  every `/api/v1/*` request (`/swagger/*` is exempt). If `API_KEY` is
  unset, the server logs a startup warning and runs open — that's the
  default in tests (`t.Setenv("API_KEY", "")`) and in a fresh `.env.example`
  copy, so don't assume auth is on without checking. There's no user
  model, sessions, or per-device revocation — rotating `API_KEY` logs out
  every client, including the phone.
- **`Image.storagePath` is kept in sync by the processor**, via
  `repo.UpdateImagePath` (matched by the *old* `storagePath`, so it must
  still equal what `service.UploadImage`/`BulkUploadImages` wrote). If you
  change how upload paths are built, update `UpdateImagePath`'s filter
  logic too or the match will silently stop finding records.
  `processor.moveFile` returns the path it actually wrote to (it may
  rename to avoid clobbering an existing file), and that's the value
  passed to `UpdateImagePath` — don't reintroduce the old bug where the
  pre-collision path got persisted instead.
- **`ENABLE_AI_CATEGORIZATION` is still a stub.** `processor.callOpenAIVision`
  never calls OpenAI even when `OPENAI_API_KEY` is set — it does filename
  keyword matching, same as the non-AI path. If asked to "fix AI
  categorization," this is a from-scratch integration, not a bug fix.
- **Tags**: `service.ParseTags` is the single source of truth for turning
  a comma-separated `tags` form value into a trimmed `[]string`, used by
  both `POST /images` and `POST /images/multiple`.
  `POST /images/bulk-upload` takes `tags` as a JSON array directly and
  doesn't go through `ParseTags`.
- **Upload validation lives in `service.supportedImageExts`** (a
  package-level map), shared by `UploadImage` and `BulkUploadImages` — add
  a new format there, not in two separate places. `UploadImage` also runs
  `filepath.Base` on the client-supplied filename before using it in any
  path — don't remove that when touching upload code.
- **Makefile targets are prefixed**: `go-run`, `go-test`, `go-test-html`,
  `docs`, `init`, `go-clean`, `build` (in
  [house_app_server/Makefile](house_app_server/Makefile)) — not the bare
  `test`/`run` names other Go projects often use.
- **`STORAGE_DIR` defaults to `E:/house_app_storage`** — a Windows path
  baked in as this developer's actual machine convention, not a
  cross-platform default. Always override it via `.env` on another OS/box.
- `.env` is gitignored; copy `house_app_server/.env.example` to get
  started. `MONGO_URI` is the only variable with no code-level default —
  the process calls `os.Exit(1)` at startup if it's unset.
- Tests in `service/` that call `UploadImage` write to `./uploads` — they
  clean up with `t.Cleanup(func() { os.RemoveAll("./uploads") })`. Keep
  that pattern for any new test that exercises upload.

## Keeping this file current

There's no `make tree`/generator script in this repo (unlike some sibling
projects) — update this file by hand when conventions or the package
layout change.
