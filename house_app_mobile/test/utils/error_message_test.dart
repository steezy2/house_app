import 'package:flutter_test/flutter_test.dart';
import 'package:house_app_mobile/api/api_client.dart';
import 'package:house_app_mobile/utils/error_message.dart';

void main() {
  group('describeApiError', () {
    test('returns the ApiConfigException message as-is', () {
      const error = ApiConfigException('No server address configured.');
      expect(describeApiError(error), 'No server address configured.');
    });

    test('uses unauthorizedMessage for a 401 ApiException when given', () {
      final error = ApiException('missing or invalid X-API-Key header', 401);
      expect(
        describeApiError(error, unauthorizedMessage: 'Check your API key.'),
        'Check your API key.',
      );
    });

    test('falls back to a generic message for a 401 with no override', () {
      final error = ApiException('missing or invalid X-API-Key header', 401);
      expect(describeApiError(error), 'The server rejected the API key.');
    });

    test('uses the server message for a non-401 ApiException', () {
      final error = ApiException('unsupported file type ".txt"', 400);
      expect(describeApiError(error), 'unsupported file type ".txt"');
    });

    test('wraps any other error as an unreachable-server message', () {
      expect(
        describeApiError(Exception('Connection refused')),
        'Could not reach the server: Exception: Connection refused',
      );
    });
  });
}
