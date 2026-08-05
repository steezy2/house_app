import 'package:flutter_test/flutter_test.dart';
import 'package:house_app_mobile/models/image_asset.dart';

void main() {
  group('ImageAsset.fromJson', () {
    test('parses a fully-populated, already-processed record', () {
      final image = ImageAsset.fromJson({
        'id': '507f1f77bcf86cd799439011',
        'filename': 'beach.jpg',
        'originalPath': 'beach.jpg',
        'storagePath':
            'E:/house_app_storage/2026/01/travel/travel_20260101_103000_beach.jpg',
        'size': 102400,
        'contentType': 'image/jpeg',
        'tags': ['vacation', 'beach'],
        'category': 'travel',
        'createdAt': '2026-01-01T10:30:00Z',
        'processedAt': '2026-01-01T10:35:00Z',
      });

      expect(image.id, '507f1f77bcf86cd799439011');
      expect(image.filename, 'beach.jpg');
      expect(image.size, 102400);
      expect(image.contentType, 'image/jpeg');
      expect(image.tags, ['vacation', 'beach']);
      expect(image.category, 'travel');
      expect(image.isProcessed, isTrue);
    });

    test(
      'treats missing tags/category and a Go zero-time processedAt as unset',
      () {
        // house_app_server's Image.ProcessedAt is a non-pointer time.Time
        // with `omitempty`, which Go's encoding/json does not actually
        // omit — it serializes the zero value instead. A freshly-uploaded,
        // not-yet-processed image looks like this on the wire.
        final image = ImageAsset.fromJson({
          'id': 'i1',
          'filename': 'photo.jpg',
          'originalPath': 'photo.jpg',
          'storagePath': './uploads/1730000000_photo.jpg',
          'size': 1024,
          'contentType': 'image/jpeg',
          'createdAt': '2026-01-01T10:30:00Z',
          'processedAt': '0001-01-01T00:00:00Z',
        });

        expect(image.tags, isEmpty);
        expect(image.category, isNull);
        expect(image.processedAt, isNull);
        expect(image.isProcessed, isFalse);
      },
    );
  });
}
