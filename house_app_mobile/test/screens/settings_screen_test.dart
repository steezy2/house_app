import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:house_app_mobile/api/api.dart';
import 'package:house_app_mobile/screens/settings/settings_screen.dart';
import 'package:house_app_mobile/state/settings_state.dart';
import 'package:provider/provider.dart';

import '../fakes/fake_settings_storage.dart';

Future<SettingsState> _pumpSettingsScreen(
  WidgetTester tester, {
  required FakeSettingsStorage storage,
  required http.Client httpClient,
}) async {
  final settingsState = SettingsState(storage);
  await settingsState.hydrate();
  final api = Api(httpClient: httpClient, settingsStorage: storage);

  await tester.pumpWidget(
    MultiProvider(
      providers: [
        Provider<Api>.value(value: api),
        ChangeNotifierProvider<SettingsState>.value(value: settingsState),
      ],
      child: const MaterialApp(home: SettingsScreen()),
    ),
  );
  await tester.pumpAndSettle();
  return settingsState;
}

void main() {
  testWidgets('pre-fills fields from previously saved settings', (
    tester,
  ) async {
    final storage = FakeSettingsStorage();
    await storage.writeBaseUrl('http://192.168.1.23:8080');
    await storage.writeApiKey('existing-key');

    await _pumpSettingsScreen(
      tester,
      storage: storage,
      httpClient: MockClient((request) async => http.Response('[]', 200)),
    );

    expect(find.text('http://192.168.1.23:8080'), findsOneWidget);
    expect(find.text('existing-key'), findsOneWidget);
  });

  testWidgets('Save persists the typed server address and API key', (
    tester,
  ) async {
    final storage = FakeSettingsStorage();
    final settingsState = await _pumpSettingsScreen(
      tester,
      storage: storage,
      httpClient: MockClient((request) async => http.Response('[]', 200)),
    );

    await tester.enterText(
      find.widgetWithText(TextField, 'Server address'),
      '192.168.1.50:8080',
    );
    await tester.enterText(
      find.widgetWithText(TextField, 'API key'),
      'new-key',
    );
    await tester.tap(find.widgetWithText(ElevatedButton, 'Save'));
    await tester.pumpAndSettle();

    expect(settingsState.baseUrl, 'http://192.168.1.50:8080');
    expect(await storage.readApiKey(), 'new-key');
    expect(find.text('Settings saved'), findsOneWidget);
  });

  testWidgets('Test Connection shows success on a 200 response', (
    tester,
  ) async {
    final storage = FakeSettingsStorage();
    await _pumpSettingsScreen(
      tester,
      storage: storage,
      httpClient: MockClient((request) async => http.Response('[]', 200)),
    );

    await tester.enterText(
      find.widgetWithText(TextField, 'Server address'),
      '192.168.1.23:8080',
    );
    await tester.tap(find.widgetWithText(OutlinedButton, 'Test Connection'));
    await tester.pumpAndSettle();

    expect(find.text('Connected successfully.'), findsOneWidget);
  });

  testWidgets('Test Connection reports a rejected API key on 401', (
    tester,
  ) async {
    final storage = FakeSettingsStorage();
    await _pumpSettingsScreen(
      tester,
      storage: storage,
      httpClient: MockClient(
        (request) async =>
            http.Response('{"error":"missing or invalid X-API-Key header"}', 401),
      ),
    );

    await tester.enterText(
      find.widgetWithText(TextField, 'Server address'),
      '192.168.1.23:8080',
    );
    await tester.tap(find.widgetWithText(OutlinedButton, 'Test Connection'));
    await tester.pumpAndSettle();

    expect(
      find.text('Connected, but the server rejected the API key.'),
      findsOneWidget,
    );
  });
}
