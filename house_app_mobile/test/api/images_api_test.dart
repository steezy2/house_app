import 'dart:convert';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:house_app_mobile/api/api_client.dart';
import 'package:house_app_mobile/api/images_api.dart';
import 'package:image_picker/image_picker.dart';

import '../test_helpers.dart';

void main() {
  group('ImagesApi.getImages', () {
    test('decodes the bare JSON array the server returns', () async {
      final storage = await configuredStorage();
      final client = ApiClient(
        httpClient: MockClient(
          (request) async => http.Response(
            '[{"id":"i1","filename":"a.jpg","originalPath":"a.jpg",'
            '"storagePath":"./uploads/a.jpg","size":100,'
            '"contentType":"image/jpeg","createdAt":"2026-01-01T00:00:00Z"}]',
            200,
          ),
        ),
        settingsStorage: storage,
      );

      final images = await ImagesApi(client).getImages();

      expect(images, hasLength(1));
      expect(images.single.filename, 'a.jpg');
    });
  });

  group('ImagesApi.uploadImages', () {
    late Directory tempDir;

    setUp(() {
      tempDir = Directory.systemTemp.createTempSync('images_api_test');
    });

    tearDown(() {
      tempDir.deleteSync(recursive: true);
    });

    test(
      'posts every file under the "files" field with comma-joined tags',
      () async {
        final file1 = File('${tempDir.path}/a.jpg')
          ..writeAsBytesSync([1, 2, 3]);
        final file2 = File('${tempDir.path}/b.jpg')
          ..writeAsBytesSync([4, 5, 6]);

        final storage = await configuredStorage(apiKey: 'secret-key');

        late http.BaseRequest captured;
        final client = ApiClient(
          httpClient: MockClient.streaming((request, bodyStream) async {
            captured = request;
            await bodyStream.drain<void>();
            return http.StreamedResponse(
              Stream.value(
                utf8.encode(
                  '{"totalImages":2,"successfulCount":2,"failedCount":0}',
                ),
              ),
              200,
            );
          }),
          settingsStorage: storage,
        );

        final result = await ImagesApi(client).uploadImages(
          [XFile(file1.path), XFile(file2.path)],
          tags: ['family', '2026'],
        );

        expect(result.successfulCount, 2);
        expect(captured, isA<http.MultipartRequest>());
        final multipart = captured as http.MultipartRequest;
        expect(
          multipart.url.toString(),
          'http://192.168.1.23:8080/api/v1/images/multiple',
        );
        expect(multipart.headers['X-API-Key'], 'secret-key');
        expect(multipart.fields['tags'], 'family,2026');
        expect(multipart.files, hasLength(2));
        expect(multipart.files.every((f) => f.field == 'files'), isTrue);
      },
    );

    test('omits the tags field when no tags are given', () async {
      final file = File('${tempDir.path}/a.jpg')..writeAsBytesSync([1]);
      final storage = await configuredStorage();

      late http.BaseRequest captured;
      final client = ApiClient(
        httpClient: MockClient.streaming((request, bodyStream) async {
          captured = request;
          await bodyStream.drain<void>();
          return http.StreamedResponse(
            Stream.value(
              utf8.encode(
                '{"totalImages":1,"successfulCount":1,"failedCount":0}',
              ),
            ),
            200,
          );
        }),
        settingsStorage: storage,
      );

      await ImagesApi(client).uploadImages([XFile(file.path)]);

      final multipart = captured as http.MultipartRequest;
      expect(multipart.fields.containsKey('tags'), isFalse);
    });
  });
}
