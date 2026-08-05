import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// Persists the configured server address and API key in the platform
/// keychain/keystore. Unlike a typical login token, both of these are
/// user-entered configuration rather than issued by the server, but they
/// live in secure storage anyway since the API key is a credential.
class SettingsStorage {
  SettingsStorage() : _storage = const FlutterSecureStorage();

  final FlutterSecureStorage _storage;
  static const _baseUrlKey = 'house_app_base_url';
  static const _apiKeyKey = 'house_app_api_key';

  Future<String?> readBaseUrl() => _storage.read(key: _baseUrlKey);

  Future<String?> readApiKey() => _storage.read(key: _apiKeyKey);

  Future<void> writeBaseUrl(String value) =>
      _storage.write(key: _baseUrlKey, value: value);

  Future<void> writeApiKey(String value) =>
      _storage.write(key: _apiKeyKey, value: value);
}
