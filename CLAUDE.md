# CLAUDE.md

Project navigation notes for Claude Code. Human-facing docs live in
[README.md](README.md), [house_app_server/README.md](house_app_server/README.md),
[house_app_server/API_EXAMPLES.md](house_app_server/API_EXAMPLES.md),
[house_app_mobile/README.md](house_app_mobile/README.md), and
[IMAGE_PROCESSING.md](IMAGE_PROCESSING.md) — this file is a faster map for
finding things and a list of non-obvious traps, not a replacement for them.

## What this is

A personal/home photo-library system: a Go API
([house_app_server](house_app_server)) plus a Flutter phone client
([house_app_mobile](house_app_mobile)) that's the primary intended way to
use it — pick photos on your phone and upload them in a tap while on the
home network. There's no web frontend; anything not from the phone app
goes through curl or the generated Swagger UI. The server separately runs
a background worker that reorganizes `./uploads` into a dated, categorized
folder tree. The server started life as a bookmark-manager
(`bookmark_server`) — see [MIGRATION_SUMMARY.md](MIGRATION_SUMMARY.md) for
that rewrite's history if you find a stray reference to bookmarks/scraping
anywhere.

## Where things live

- `house_app_server/`: the Go API (single module, `house-app`). Layered
  `clients` (Mongo) -> `models` (domain types + repo interface) ->
  `service` (upload/business logic) -> `controller` (HTTP routes + swag
  annotations). Composition root: `main.go`.
  - `processor/image_processor.go`: independent background loop, started
    from `main.go` alongside (not through) the HTTP server. Polls
    `UPLOAD_DIR`, extracts EXIF, categorizes, renames, and moves files into
    `STORAGE_DIR/YYYY/MM/category/`.
  - `docs/`: generated Swagger/OpenAPI output (`docs.go`, `swagger.json`,
    `swagger.yaml`). Regenerate with `make docs` after touching swag
    annotations in `controller/api.go` or `main.go` — don't hand-edit.
- `house_app_mobile/`: the Flutter client (package `house_app_mobile`).
  Layout mirrors `InsuranceMarketplace/apps/mobile` (`api/`, `models/`,
  `screens/`, `state/`, `theme/`, `widgets/`), deliberately without that
  app's `go_router` or login flow — see its own CLAUDE-equivalent notes
  below and [house_app_mobile/README.md](house_app_mobile/README.md).
- Root-level `*.md` files are the human docs; `MIGRATION_SUMMARY.md` is a
  point-in-time record of the bookmark→house-app rewrite, not living
  documentation — don't "fix" it to match current behavior.

## Server: key conventions and known traps

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

## Mobile: key conventions and known traps

- **`ApiClient` reads `SettingsStorage` fresh on every request**, not a
  cached value — same pattern as the sibling Insurance Marketplace app's
  `ApiClient` reading `TokenStorage` independently of `AuthState`. If you
  add a new API-backed screen, don't thread `baseUrl`/`apiKey` through as
  parameters; let `ApiClient` resolve them itself.
- **No thumbnails in the Library tab, on purpose.** `house_app_server` has
  no endpoint to fetch raw image bytes, so `ImageAsset` only carries
  metadata. Don't add an `Image.network(...)` call against `storagePath`
  — that's a server filesystem path, not a URL, and won't resolve.
- **Plain HTTP to a LAN address requires platform config** — see
  `house_app_mobile/README.md`'s "Plain HTTP to a LAN address" section.
  If a fresh device/emulator can't reach the server despite correct
  Settings, check that `usesCleartextTraffic`
  (`android/app/src/main/AndroidManifest.xml`) and
  `NSAllowsLocalNetworking` (`ios/Runner/Info.plist`) weren't reverted by
  a `flutter create`/upgrade regenerating those files.
- **`SettingsState.normalizeBaseUrl`** is the only place that should touch
  user-typed server addresses — it defaults a bare `192.168.1.23:8080` to
  `http://`, and strips trailing slashes so `'$baseUrl$path'`
  concatenation in `ApiClient._uri` doesn't produce `//api`. Route new
  settings-editing UI through `SettingsState.save`, not `SettingsStorage`
  directly, or you'll bypass this.
- **No `go_router`, unlike the Insurance Marketplace app** — there's no
  auth-gated routing here (a single shared `API_KEY`, no per-user
  sessions), so `widgets/bottom_nav_shell.dart` is a plain `IndexedStack`.
  Don't add `go_router` back in without an actual reason (deep links,
  route guards) to justify it.
- Tests avoid real `flutter_secure_storage` (its platform channels aren't
  available under plain `flutter test`) via `test/fakes/fake_settings_storage.dart`,
  and avoid real network calls via `package:http/testing.dart`'s
  `MockClient`/`MockClient.streaming` — the latter is required (not the
  plain `MockClient`) for asserting on multipart upload fields/files,
  since only the streaming handler receives the original
  `MultipartRequest` instance rather than a reconstructed plain `Request`.

## Keeping this file current

There's no `make tree`/generator script in this repo (unlike some sibling
projects) — update this file by hand when conventions or the package
layout change.
