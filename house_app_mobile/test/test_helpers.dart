import 'fakes/fake_settings_storage.dart';

/// Builds a [FakeSettingsStorage] pre-populated with a base URL (and
/// optionally an API key) — the setup nearly every ApiClient/ImagesApi
/// test needs before it can make a request.
Future<FakeSettingsStorage> configuredStorage({
  String baseUrl = 'http://192.168.1.23:8080',
  String? apiKey,
}) async {
  final storage = FakeSettingsStorage();
  await storage.writeBaseUrl(baseUrl);
  if (apiKey != null) {
    await storage.writeApiKey(apiKey);
  }
  return storage;
}
