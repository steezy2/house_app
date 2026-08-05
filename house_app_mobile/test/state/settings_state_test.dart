import 'package:flutter_test/flutter_test.dart';
import 'package:house_app_mobile/state/settings_state.dart';

import '../fakes/fake_settings_storage.dart';
import '../test_helpers.dart';

void main() {
  group('SettingsState', () {
    test('hydrate starts unconfigured when nothing is stored', () async {
      final state = SettingsState(FakeSettingsStorage());

      await state.hydrate();

      expect(state.loading, isFalse);
      expect(state.isConfigured, isFalse);
      expect(state.baseUrl, isNull);
    });

    test('hydrate loads previously saved values', () async {
      final storage = await configuredStorage(apiKey: 'secret');
      final state = SettingsState(storage);

      await state.hydrate();

      expect(state.isConfigured, isTrue);
      expect(state.baseUrl, 'http://192.168.1.23:8080');
      expect(state.apiKey, 'secret');
    });

    test('save normalizes the URL, persists, and notifies listeners', () async {
      final storage = FakeSettingsStorage();
      final state = SettingsState(storage);
      await state.hydrate();
      var notified = false;
      state.addListener(() => notified = true);

      await state.save(baseUrl: '192.168.1.23:8080', apiKey: 'secret');

      expect(notified, isTrue);
      expect(state.baseUrl, 'http://192.168.1.23:8080');
      expect(await storage.readBaseUrl(), 'http://192.168.1.23:8080');
      expect(await storage.readApiKey(), 'secret');
    });
  });

  group('SettingsState.normalizeBaseUrl', () {
    test('defaults to http:// when no scheme is given', () {
      expect(
        SettingsState.normalizeBaseUrl('192.168.1.23:8080'),
        'http://192.168.1.23:8080',
      );
    });

    test('preserves an explicit https scheme', () {
      expect(
        SettingsState.normalizeBaseUrl('https://example.com'),
        'https://example.com',
      );
    });

    test('strips a trailing slash', () {
      expect(
        SettingsState.normalizeBaseUrl('http://192.168.1.23:8080/'),
        'http://192.168.1.23:8080',
      );
    });

    test('trims surrounding whitespace', () {
      expect(
        SettingsState.normalizeBaseUrl('  192.168.1.23:8080  '),
        'http://192.168.1.23:8080',
      );
    });

    test('leaves an empty string empty', () {
      expect(SettingsState.normalizeBaseUrl(''), '');
    });
  });
}
