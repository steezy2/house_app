import 'package:flutter/foundation.dart';

import 'settings_storage.dart';

/// Holds the server address and API key for the rest of the app to read
/// reactively, backed by [SettingsStorage]. [ApiClient] does not depend on
/// this class directly — it reads [SettingsStorage] itself on every
/// request, the same way the sibling Insurance Marketplace app's
/// `ApiClient` reads `TokenStorage` independently of `AuthState`.
class SettingsState extends ChangeNotifier {
  SettingsState(this._storage);

  final SettingsStorage _storage;

  String? _baseUrl;
  String? _apiKey;
  bool _loading = true;

  String? get baseUrl => _baseUrl;
  String? get apiKey => _apiKey;
  bool get loading => _loading;
  bool get isConfigured => _baseUrl != null && _baseUrl!.isNotEmpty;

  /// Call once at app start.
  Future<void> hydrate() async {
    _baseUrl = await _storage.readBaseUrl();
    _apiKey = await _storage.readApiKey();
    _loading = false;
    notifyListeners();
  }

  Future<void> save({required String baseUrl, required String apiKey}) async {
    final normalized = normalizeBaseUrl(baseUrl);
    await _storage.writeBaseUrl(normalized);
    await _storage.writeApiKey(apiKey);
    _baseUrl = normalized;
    _apiKey = apiKey;
    notifyListeners();
  }

  /// Users on a home network naturally type `192.168.1.23:8080`, not a full
  /// URL — default to `http://` when no scheme is given, and drop any
  /// trailing slash so `$baseUrl$path` concatenation in [ApiClient] doesn't
  /// produce a double slash.
  static String normalizeBaseUrl(String value) {
    var url = value.trim();
    if (url.isEmpty) return url;
    if (!url.contains('://')) url = 'http://$url';
    while (url.endsWith('/')) {
      url = url.substring(0, url.length - 1);
    }
    return url;
  }
}
