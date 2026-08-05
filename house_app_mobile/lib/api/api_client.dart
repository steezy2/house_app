import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:image_picker/image_picker.dart';

import '../state/settings_storage.dart';

/// Thrown for any non-2xx response, carrying the backend's `{"error": "..."}`
/// message when present.
class ApiException implements Exception {
  ApiException(this.message, this.statusCode);

  final String message;
  final int statusCode;

  @override
  String toString() => message;
}

/// Thrown when a request is attempted before a server address has been
/// configured in Settings.
class ApiConfigException implements Exception {
  const ApiConfigException(this.message);

  final String message;

  @override
  String toString() => message;
}

/// Low-level HTTP wrapper: reads the current server address/API key from
/// [SettingsStorage] on every call (so a change in Settings takes effect
/// immediately, with no need to rebuild this client), injects the
/// `X-API-Key` header, and decodes JSON / throws [ApiException] on
/// failure — mirrors house_app_server's `controller.APIKeyMiddleware`
/// contract on the other end.
class ApiClient {
  ApiClient({http.Client? httpClient, SettingsStorage? settingsStorage})
    : _http = httpClient ?? http.Client(),
      _settingsStorage = settingsStorage ?? SettingsStorage();

  final http.Client _http;
  final SettingsStorage _settingsStorage;

  Future<Uri> _uri(String path, [Map<String, String>? query]) async {
    final baseUrl = await _settingsStorage.readBaseUrl();
    if (baseUrl == null || baseUrl.isEmpty) {
      throw const ApiConfigException(
        'No server address configured. Set one in Settings first.',
      );
    }
    final uri = Uri.parse('$baseUrl$path');
    if (query == null || query.isEmpty) return uri;
    return uri.replace(queryParameters: query);
  }

  Future<Map<String, String>> _headers() async {
    final headers = <String, String>{};
    final apiKey = await _settingsStorage.readApiKey();
    if (apiKey != null && apiKey.isNotEmpty) {
      headers['X-API-Key'] = apiKey;
    }
    return headers;
  }

  Future<dynamic> get(String path, {Map<String, String>? query}) async {
    final response = await _http.get(
      await _uri(path, query),
      headers: await _headers(),
    );
    return _decode(response);
  }

  /// Uploads [files] as a single multipart POST to [path] under
  /// [fileField] (repeated once per file, matching how Gin's
  /// `form.File["files"]` expects a multi-file field), attaching [tags] as
  /// a single comma-separated form value if given.
  Future<dynamic> postMultipart(
    String path, {
    required String fileField,
    required List<XFile> files,
    String? tags,
  }) async {
    final request = http.MultipartRequest('POST', await _uri(path));
    request.headers.addAll(await _headers());
    if (tags != null && tags.isNotEmpty) {
      request.fields['tags'] = tags;
    }
    for (final file in files) {
      request.files.add(
        await http.MultipartFile.fromPath(
          fileField,
          file.path,
          filename: file.name,
        ),
      );
    }

    final streamedResponse = await _http.send(request);
    final response = await http.Response.fromStream(streamedResponse);
    return _decode(response);
  }

  dynamic _decode(http.Response response) {
    final isSuccess = response.statusCode >= 200 && response.statusCode < 300;

    dynamic decoded;
    if (response.body.isNotEmpty) {
      try {
        decoded = jsonDecode(response.body);
      } catch (_) {
        decoded = null;
      }
    }

    if (!isSuccess) {
      final message = decoded is Map<String, dynamic>
          ? decoded['error'] as String?
          : null;
      throw ApiException(
        message ?? 'Request failed (${response.statusCode})',
        response.statusCode,
      );
    }

    return decoded;
  }
}
