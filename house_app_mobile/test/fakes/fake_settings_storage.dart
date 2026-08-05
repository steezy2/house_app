import 'package:house_app_mobile/state/settings_storage.dart';

/// In-memory [SettingsStorage] for tests — avoids `flutter_secure_storage`'s
/// platform channels, which aren't available under plain `flutter test`.
class FakeSettingsStorage extends SettingsStorage {
  String? _baseUrl;
  String? _apiKey;

  @override
  Future<String?> readBaseUrl() async => _baseUrl;

  @override
  Future<String?> readApiKey() async => _apiKey;

  @override
  Future<void> writeBaseUrl(String value) async => _baseUrl = value;

  @override
  Future<void> writeApiKey(String value) async => _apiKey = value;
}
