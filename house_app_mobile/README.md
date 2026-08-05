# House App Mobile

Flutter client for [house_app_server](../house_app_server): pick photos on
your phone and upload them to your home House App server in one tap. This
is the primary intended client for that API — see
[house_app_server/API_EXAMPLES.md](../house_app_server/API_EXAMPLES.md) for
the equivalent curl calls.

## What it does

- **Upload tab**: select any number of photos from your gallery, optionally
  tag them, and upload them all in a single request to
  `POST /api/v1/images/multiple`.
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

## Testing

```bash
flutter test
```

Tests use `package:http/testing.dart`'s `MockClient` and an in-memory
`FakeSettingsStorage` (see `test/fakes/`) — no real network calls or
platform-channel-backed secure storage, so they run under plain
`flutter test`.

## Structure

Mirrors the layout of the sibling
[InsuranceMarketplace `apps/mobile`](../../InsuranceMarketplace/apps/mobile)
app (`api/`, `models/`, `screens/`, `state/`, `theme/`, `widgets/`), minus
anything that app needed but this one doesn't — no `go_router` (there are
no auth-gated routes here, so a plain `IndexedStack` bottom nav is enough)
and no login flow (auth here is a single shared `API_KEY`, not per-user
accounts).
