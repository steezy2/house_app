import 'package:flutter_test/flutter_test.dart';
import 'package:house_app_mobile/models/bulk_upload_result.dart';
import 'package:house_app_mobile/services/media_library.dart';
import 'package:house_app_mobile/utils/upload_results.dart';
import 'package:image_picker/image_picker.dart';

BulkUploadResult _result({List<String> failedFiles = const []}) =>
    BulkUploadResult(
      totalImages: 2,
      successfulCount: 2 - failedFiles.length,
      failedCount: failedFiles.length,
      failedFiles: failedFiles,
    );

void main() {
  final batch = [
    ResolvedAsset(id: 'id-a', file: XFile('photo.jpg'), sizeBytes: 1),
    ResolvedAsset(id: 'id-b', file: XFile('my_photo.jpg'), sizeBytes: 1),
  ];

  test('returns every id when the whole batch succeeded', () {
    expect(succeededAssetIds(batch, _result()), ['id-a', 'id-b']);
  });

  test('leaves out assets the server reported as failed', () {
    expect(
      succeededAssetIds(
        batch,
        _result(failedFiles: ['my_photo.jpg: unsupported file type']),
      ),
      ['id-a'],
    );
  });
}
