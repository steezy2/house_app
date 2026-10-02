import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:house_app_mobile/api/api.dart';
import 'package:house_app_mobile/screens/upload/upload_screen.dart';
import 'package:house_app_mobile/services/media_library.dart';
import 'package:house_app_mobile/services/uploaded_assets_store.dart';
import 'package:house_app_mobile/state/settings_state.dart';
import 'package:image_picker/image_picker.dart';
import 'package:provider/provider.dart';

import '../fakes/fake_uploaded_assets_store.dart';
import '../test_helpers.dart';

/// Note: this deliberately never exercises a real upload through
/// `ApiClient.postMultipart` — `http.MultipartFile.fromPath`'s real file
/// I/O hangs indefinitely when awaited directly inside a `testWidgets`
/// body in this environment (reproduced and confirmed with a minimal
/// repro; a plain `test()` with the identical file read completes
/// instantly). The actual multipart upload mechanics are covered at the
/// `test()` level in `test/api/images_api_test.dart`, which isn't
/// affected. These tests only exercise the parts of the mass-upload flow
/// that resolve before any file is read.
class FakeMediaLibrary implements MediaLibrary {
  FakeMediaLibrary({required this.accessResult, this.assets = const []});

  final MediaAccessResult accessResult;
  final List<ResolvedAsset> assets;

  @override
  Future<MediaAccessResult> requestAccess() async => accessResult;

  @override
  Future<List<ResolvedAsset>> getAllAssets({
    Set<String> exclude = const {},
  }) async => assets.where((a) => !exclude.contains(a.id)).toList();
}

Future<void> _pumpUploadScreen(
  WidgetTester tester, {
  required MediaLibrary mediaLibrary,
  required http.Client httpClient,
  UploadedAssetsStore? uploadedAssets,
}) async {
  final storage = await configuredStorage(apiKey: 'secret-key');
  final settingsState = SettingsState(storage);
  await settingsState.hydrate();
  final api = Api(httpClient: httpClient, settingsStorage: storage);

  await tester.pumpWidget(
    MultiProvider(
      providers: [
        Provider<Api>.value(value: api),
        ChangeNotifierProvider<SettingsState>.value(value: settingsState),
      ],
      child: MaterialApp(
        home: UploadScreen(
          mediaLibrary: mediaLibrary,
          uploadedAssets: uploadedAssets ?? FakeUploadedAssetsStore(),
        ),
      ),
    ),
  );
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('shows an error when photo library access is denied', (
    tester,
  ) async {
    await _pumpUploadScreen(
      tester,
      mediaLibrary: FakeMediaLibrary(accessResult: MediaAccessResult.denied),
      httpClient: MockClient((request) async => http.Response('[]', 200)),
    );

    await tester.tap(find.text('Upload All Photos & Videos'));
    await tester.pumpAndSettle();

    expect(
      find.textContaining('Photo library access was denied'),
      findsOneWidget,
    );
  });

  testWidgets('shows a message when no assets are found', (tester) async {
    await _pumpUploadScreen(
      tester,
      mediaLibrary: FakeMediaLibrary(accessResult: MediaAccessResult.full),
      httpClient: MockClient((request) async => http.Response('[]', 200)),
    );

    await tester.tap(find.text('Upload All Photos & Videos'));
    await tester.pumpAndSettle();

    expect(
      find.text('No photos or videos found on this device.'),
      findsOneWidget,
    );
  });

  testWidgets('confirmation dialog shows the asset count and cancels cleanly', (
    tester,
  ) async {
    final mediaLibrary = FakeMediaLibrary(
      accessResult: MediaAccessResult.full,
      // Paths don't need to exist — canceling never reads them.
      assets: [
        ResolvedAsset(id: 'id-a', file: XFile('a.jpg'), sizeBytes: 3),
        ResolvedAsset(id: 'id-b', file: XFile('b.mp4'), sizeBytes: 3),
      ],
    );

    await _pumpUploadScreen(
      tester,
      mediaLibrary: mediaLibrary,
      httpClient: MockClient(
        (request) async => fail('should not call the network on cancel'),
      ),
    );

    await tester.tap(find.text('Upload All Photos & Videos'));
    await tester.pumpAndSettle();

    expect(find.text('Upload everything?'), findsOneWidget);
    expect(find.textContaining('all 2 photos and videos'), findsOneWidget);

    await tester.tap(find.widgetWithText(TextButton, 'Cancel'));
    await tester.pumpAndSettle();

    expect(find.text('Upload everything?'), findsNothing);
    expect(find.text('Uploading...'), findsNothing);
  });

  testWidgets(
    'warns about limited iOS photo access in the confirmation dialog',
    (tester) async {
      final mediaLibrary = FakeMediaLibrary(
        accessResult: MediaAccessResult.limited,
        assets: [ResolvedAsset(id: 'id-a', file: XFile('a.jpg'), sizeBytes: 3)],
      );

      await _pumpUploadScreen(
        tester,
        mediaLibrary: mediaLibrary,
        httpClient: MockClient(
          (request) async => fail('should not call the network on cancel'),
        ),
      );

      await tester.tap(find.text('Upload All Photos & Videos'));
      await tester.pumpAndSettle();

      expect(find.textContaining('Limited Photos'), findsOneWidget);

      await tester.tap(find.widgetWithText(TextButton, 'Cancel'));
      await tester.pumpAndSettle();
    },
  );

  testWidgets('offers only assets that were not uploaded before', (
    tester,
  ) async {
    await _pumpUploadScreen(
      tester,
      mediaLibrary: FakeMediaLibrary(
        accessResult: MediaAccessResult.full,
        assets: [
          ResolvedAsset(id: 'id-a', file: XFile('a.jpg'), sizeBytes: 3),
          ResolvedAsset(id: 'id-b', file: XFile('b.mp4'), sizeBytes: 3),
        ],
      ),
      httpClient: MockClient(
        (request) async => fail('should not call the network on cancel'),
      ),
      uploadedAssets: FakeUploadedAssetsStore({'id-a'}),
    );

    await tester.tap(find.text('Upload All Photos & Videos'));
    await tester.pumpAndSettle();

    expect(find.textContaining('1 new photo or video'), findsOneWidget);

    await tester.tap(find.widgetWithText(TextButton, 'Cancel'));
    await tester.pumpAndSettle();
  });

  testWidgets('says so when everything is already uploaded', (tester) async {
    await _pumpUploadScreen(
      tester,
      mediaLibrary: FakeMediaLibrary(
        accessResult: MediaAccessResult.full,
        assets: [ResolvedAsset(id: 'id-a', file: XFile('a.jpg'), sizeBytes: 3)],
      ),
      httpClient: MockClient((request) async => fail('nothing to upload')),
      uploadedAssets: FakeUploadedAssetsStore({'id-a'}),
    );

    await tester.tap(find.text('Upload All Photos & Videos'));
    await tester.pumpAndSettle();

    expect(
      find.text('Everything on this device is already uploaded.'),
      findsOneWidget,
    );
  });
}
