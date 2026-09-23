# House App Mobile

Flutter client for [house_app_server](../house_app_server): pick photos on
your phone and upload them to your home House App server in one tap. This
is the primary intended client for that API — see
[house_app_server/API_EXAMPLES.md](../house_app_server/API_EXAMPLES.md) for
the equivalent curl calls.

## What it does

- **Upload tab**:
  - **Select Photos**: pick any number of photos from your gallery
    (`image_picker`'s picker UI, no special permission needed) and upload
    them in one request to `POST /api/v1/images/multiple`.
  - **Upload All Photos & Videos**: enumerates *every* photo and video
    already on the device (via `photo_manager`'s full-library access —
    a broader, more visible OS permission than the picker above needs)
    and uploads them all, split into batches of ≤20 files or ≤150MB
    (whichever comes first) so a single request doesn't try to send an
    unbounded amount of data. Shows a confirmation dialog with the asset
    count before starting, and a progress bar during upload. See
    `lib/services/media_library.dart` and `lib/utils/batching.dart`.
    **This does not delete anything from the device** — see
    [Future Enhancements](../README.md#future-enhancements) for the
    planned (not yet built) delete-after-verified-backup step.
- **Library tab**: lists everything uploaded so far (filename, tags,
  category, upload time), pull-to-refresh via `GET /api/v1/images`. This is
  metadata only — the server has no endpoint to serve image bytes back, so
  there are no thumbnails here (see house_app's root `CLAUDE.md`).
- **Settings tab**: set the server's address on your home network (e.g.
  `192.168.1.23:8080`) and the `X-API-Key` value, if the server has one
  configured. Includes a "Test Connection" button. Both are stored in the
  platform keychain/keystore, not compiled in — this app is meant to be
  pointed at whatever host you're running the server on, which can change.

## Running it

```bash
flutter pub get
flutter run
```

You'll need the House App server reachable on the same network as your
phone/emulator — see
[house_app_server/README.md](../house_app_server/README.md) to get it
running, then enter its address in this app's Settings tab.

### Plain HTTP to a LAN address

house_app_server doesn't run TLS — it's a home server, not a public one —
so this app talks to it over plain `http://`. Both platforms block that by
default:

- **Android**: `android:usesCleartextTraffic="true"` is set in
  `android/app/src/main/AndroidManifest.xml`.
- **iOS**: `NSAppTransportSecurity` → `NSAllowsLocalNetworking` is set in
  `ios/Runner/Info.plist`, scoped to local-network addresses rather than
  disabling ATS globally.

If you ever put the server behind real TLS, these become unnecessary but
harmless.

### Full photo library access

"Upload All" needs a broader permission than picking individual photos —
Android's `READ_MEDIA_IMAGES`/`READ_MEDIA_VIDEO` (declared explicitly in
`AndroidManifest.xml`; `photo_manager`'s own manifest only covers the
legacy `READ_EXTERNAL_STORAGE` up to API 32) and iOS's full Photos access
via the existing `NSPhotoLibraryUsageDescription`. On iOS, a user can grant
**Limited Photos** access instead of full access — the app detects this
(`MediaAccessResult.limited`) and says so in the confirmation dialog,
rather than silently uploading a partial library while claiming
"everything."

## Testing

```bash
flutter test
```

Tests use `package:http/testing.dart`'s `MockClient` and an in-memory
`FakeSettingsStorage` (see `test/fakes/`) — no real network calls or
platform-channel-backed secure storage, so they run under plain
`flutter test`.

> **Known test-environment gotcha**: awaiting `http.MultipartFile.fromPath`
> (real file I/O) directly inside a `testWidgets` body hangs indefinitely
> in this environment, even though the identical call in a plain `test()`
> resolves instantly (reproduced and confirmed — see
> `test/screens/upload_screen_test.dart`'s file comment). Multipart upload
> mechanics are tested at the `test()` level in
> `test/api/images_api_test.dart`; widget tests that touch the upload flow
> avoid ever letting a real file reach `postMultipart`.

## Structure

Mirrors the layout of the sibling
[InsuranceMarketplace `apps/mobile`](../../InsuranceMarketplace/apps/mobile)
app (`api/`, `models/`, `screens/`, `state/`, `theme/`, `widgets/`), minus
anything that app needed but this one doesn't — no `go_router` (there are
no auth-gated routes here, so a plain `IndexedStack` bottom nav is enough)
and no login flow (auth here is a single shared `API_KEY`, not per-user
accounts).
