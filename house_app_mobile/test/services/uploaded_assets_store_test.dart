import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:house_app_mobile/services/uploaded_assets_store.dart';

void main() {
  late Directory dir;

  setUp(() async => dir = await Directory.systemTemp.createTemp('uploaded'));
  tearDown(() async => dir.delete(recursive: true));

  UploadedAssetsStore newStore() =>
      UploadedAssetsStore(directory: () async => dir);

  test('starts empty', () async {
    expect(await newStore().load(), isEmpty);
  });

  test('remembers ids across instances and accumulates', () async {
    await newStore().addAll(['id-a', 'id-b']);
    await newStore().addAll(['id-c']);

    expect(await newStore().load(), {'id-a', 'id-b', 'id-c'});
  });
}
