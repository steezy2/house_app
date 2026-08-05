import 'package:flutter_test/flutter_test.dart';
import 'package:house_app_mobile/models/bulk_upload_result.dart';

void main() {
  group('BulkUploadResult.fromJson', () {
    test('parses a fully successful result', () {
      final result = BulkUploadResult.fromJson({
        'totalImages': 3,
        'successfulCount': 3,
        'failedCount': 0,
      });

      expect(result.totalImages, 3);
      expect(result.isFullSuccess, isTrue);
      expect(result.failedFiles, isEmpty);
    });

    test('parses a partially failed result', () {
      final result = BulkUploadResult.fromJson({
        'totalImages': 2,
        'successfulCount': 1,
        'failedCount': 1,
        'failedFiles': ['bad.txt: unsupported file type ".txt"'],
      });

      expect(result.isFullSuccess, isFalse);
      expect(result.failedFiles, ['bad.txt: unsupported file type ".txt"']);
    });

    test('an empty upload (0 files) is not treated as a success', () {
      final result = BulkUploadResult.fromJson({
        'totalImages': 0,
        'successfulCount': 0,
        'failedCount': 0,
      });

      expect(result.isFullSuccess, isFalse);
    });
  });
}
