import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:house_app_mobile/api/api_client.dart';

import '../fakes/fake_settings_storage.dart';
import '../test_helpers.dart';

void main() {
  group('ApiClient', () {
    test('throws ApiConfigException when no base URL is configured', () {
      final client = ApiClient(
        httpClient: MockClient(
          (request) async => fail('should not call the network'),
        ),
        settingsStorage: FakeSettingsStorage(),
      );

      expect(
        () => client.get('/api/v1/images'),
        throwsA(isA<ApiConfigException>()),
      );
    });

    test('sends the configured X-API-Key header to the configured URL', () async {
      final storage = await configuredStorage(apiKey: 'secret-key');
      late http.Request captured;
      final client = ApiClient(
        httpClient: MockClient((request) async {
          captured = request;
          return http.Response('[]', 200);
        }),
        settingsStorage: storage,
      );

      await client.get('/api/v1/images');

      expect(
        captured.url.toString(),
        'http://192.168.1.23:8080/api/v1/images',
      );
      expect(captured.headers['X-API-Key'], 'secret-key');
    });

    test('omits X-API-Key when no key is configured', () async {
      final storage = await configuredStorage();
      late http.Request captured;
      final client = ApiClient(
        httpClient: MockClient((request) async {
          captured = request;
          return http.Response('[]', 200);
        }),
        settingsStorage: storage,
      );

      await client.get('/api/v1/images');

      expect(captured.headers.containsKey('X-API-Key'), isFalse);
    });

    test('applies query parameters', () async {
      final storage = await configuredStorage();
      late http.Request captured;
      final client = ApiClient(
        httpClient: MockClient((request) async {
          captured = request;
          return http.Response('[]', 200);
        }),
        settingsStorage: storage,
      );

      await client.get('/api/v1/images/search', query: {'q': 'vacation'});

      expect(captured.url.queryParameters['q'], 'vacation');
    });

    test('throws ApiException with the server message on failure', () async {
      final storage = await configuredStorage();
      final client = ApiClient(
        httpClient: MockClient(
          (request) async => http.Response(
            '{"error":"missing or invalid X-API-Key header"}',
            401,
          ),
        ),
        settingsStorage: storage,
      );

      expect(
        () => client.get('/api/v1/images'),
        throwsA(
          isA<ApiException>()
              .having((e) => e.statusCode, 'statusCode', 401)
              .having(
                (e) => e.message,
                'message',
                'missing or invalid X-API-Key header',
              ),
        ),
      );
    });

    test('falls back to a generic message when the body has none', () async {
      final storage = await configuredStorage();
      final client = ApiClient(
        httpClient: MockClient((request) async => http.Response('', 500)),
        settingsStorage: storage,
      );

      expect(
        () => client.get('/api/v1/images'),
        throwsA(
          isA<ApiException>().having(
            (e) => e.message,
            'message',
            'Request failed (500)',
          ),
        ),
      );
    });
  });
}
