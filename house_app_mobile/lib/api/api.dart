import 'package:http/http.dart' as http;

import 'api_client.dart';
import 'images_api.dart';
import '../state/settings_storage.dart';

/// Aggregates every domain API behind a single shared [ApiClient]. Tests
/// can inject a fake [httpClient]/[settingsStorage] to avoid real network
/// calls and platform-channel-backed secure storage.
class Api {
  factory Api({http.Client? httpClient, SettingsStorage? settingsStorage}) {
    final storage = settingsStorage ?? SettingsStorage();
    final client = ApiClient(
      httpClient: httpClient,
      settingsStorage: storage,
    );
    return Api._(client: client);
  }

  Api._({required this.client}) : images = ImagesApi(client);

  final ApiClient client;
  final ImagesApi images;
}
