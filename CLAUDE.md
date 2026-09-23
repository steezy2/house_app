# CLAUDE.md

Project navigation notes for Claude Code. Human-facing docs live in
[README.md](README.md), [house_app_server/README.md](house_app_server/README.md),
[house_app_server/API_EXAMPLES.md](house_app_server/API_EXAMPLES.md),
[house_app_mobile/README.md](house_app_mobile/README.md), and
[IMAGE_PROCESSING.md](IMAGE_PROCESSING.md) — this file is a faster map for
finding things and a list of non-obvious traps, not a replacement for them.
[CHANGELOG.md](CHANGELOG.md) is the chronological log of notable changes —
check its most recent entries before assuming current behavior; this file
isn't versioned per-change, so it can lag behind the changelog.

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

- `house_app_server/`: the Go API (single module, `house-app`, requiring
  Go 1.24+ since the `backup/` package's AWS SDK dependency bumped it from
  1.21). Layered `clients` (Mongo) -> `models` (domain types + repo
  interface) -> `service` (upload/business logic) -> `controller` (HTTP
  routes + swag annotations). Composition root: `main.go`.
  - `processor/image_processor.go`: independent background loop, started
    from `main.go` alongside (not through) the HTTP server. Polls
    `UPLOAD_DIR`, extracts EXIF, categorizes, renames, moves files into
    `STORAGE_DIR/YYYY/MM/category/`, then mirrors each to every configured
    `backup.Destination`.
  - `backup/`: `Destination` interface (`Name()` + `Copy(ctx,
    relativePath, localPath)`) with `LocalDestination` (another
    drive/NAS path) and `S3Destination` (any S3-compatible bucket)
    implementations. `main.go`'s `buildBackupDestinations` assembles the
    configured list from `BACKUP_LOCAL_DIRS`/`BACKUP_S3_*` env vars.
  - `internal/mocks/`: shared `models.ImageRepository` test double used by
    `controller`, `service`, and `processor` tests — don't hand-roll
    another copy; add fields there if a test needs a new repo method
    mocked.
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
- **Upload validation lives in `models.SupportedImageExtensions` /
  `SupportedVideoExtensions`** (package-level maps in `models`, not
  `service`, so `processor.isMediaFile` can share them too) — add a new
  format there, not in per-package copies; that's exactly the drift that
  used to exist between upload validation and the processor's file-type
  check before they were unified. `service.checkSupportedExt` and
  `contentTypeForExt` both consult these sets. `UploadImage` also runs
  `filepath.Base` on the client-supplied filename before using it in any
  path — don't remove that when touching upload code.
- **Backup is additive and best-effort, never blocking.** A
  `backup.Destination` failing (drive unplugged, bucket unreachable) is
  logged as a warning in `processor.backupFile` and does not fail
  processing or affect other destinations — the primary copy in
  `STORAGE_DIR` already succeeded by the time backup runs.
  `models.Image.BackedUpTo` only lists destinations that actually
  confirmed a copy, matched by the *current* `storagePath` (same
  match-by-storagePath pattern as `UpdateImagePath`). A delete-from-phone
  feature is explicitly planned to depend on `BackedUpTo` covering every
  configured destination before treating a file as safe to remove from
  the source device — don't build that without checking this field.
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
- **`lib/services/media_library.dart` wraps `photo_manager`** for the
  Upload tab's "Upload All" full-camera-roll flow — a materially different
  permission model than `image_picker`'s picker (which needs no OS
  permission at all): full library access on Android needs
  `READ_MEDIA_IMAGES`/`READ_MEDIA_VIDEO` (declared explicitly in
  `AndroidManifest.xml`; `photo_manager`'s own manifest only covers legacy
  `READ_EXTERNAL_STORAGE` up to API 32), and on iOS can come back as
  `MediaAccessResult.limited` (the user granted only some photos) rather
  than `.full` — don't collapse that distinction, the UI is supposed to
  say so rather than silently uploading a partial library.
  `UploadScreen(mediaLibrary: ...)` is constructor-injectable specifically
  so tests can supply a `FakeMediaLibrary` instead of hitting real
  platform channels.
- **`lib/utils/batching.dart`'s `batchIndicesBySize`** caps each upload
  batch at ≤20 files or ≤150MB combined (whichever comes first) — it's a
  pure function operating on a list of byte sizes and returning index
  groups, deliberately decoupled from `photo_manager`/`image_picker`
  types, so don't thread those into it.
- **Known test-environment trap**: awaiting `http.MultipartFile.fromPath`
  (real file I/O) directly inside a `testWidgets` body hangs indefinitely
  in this environment — reproduced and confirmed with a minimal repro; the
  identical call in a plain `test()` resolves instantly. This is why
  `test/screens/upload_screen_test.dart` never lets a real file reach
  `postMultipart`, and why actual multipart-upload mechanics are tested at
  the `test()` level in `test/api/images_api_test.dart` instead. If a
  widget test involving file upload seems to hang with no error, this is
  almost certainly why — don't spend time re-debugging it, restructure the
  test to avoid real file I/O inside `testWidgets` instead.

## Keeping this file current

There's no `make tree`/generator script in this repo (unlike some sibling
projects) — update this file by hand when conventions or the package
layout change.
