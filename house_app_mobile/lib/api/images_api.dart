import 'package:image_picker/image_picker.dart';

import 'api_client.dart';
import '../models/bulk_upload_result.dart';
import '../models/image_asset.dart';

/// house_app_server's list endpoints return a bare JSON array (not the
/// `{"key": [...]}` envelope some other backends use), so this decodes
/// straight from `List<dynamic>`.
class ImagesApi {
  ImagesApi(this._client);

  final ApiClient _client;

  Future<List<ImageAsset>> getImages() async {
    final json = await _client.get('/api/v1/images');
    return _parseList(json);
  }

  Future<List<ImageAsset>> searchImages(String query) async {
    final json = await _client.get(
      '/api/v1/images/search',
      query: {'q': query},
    );
    return _parseList(json);
  }

  /// Uploads every file in [files] in a single request to
  /// `POST /images/multiple` — the endpoint a "select photos, tap upload"
  /// button should call, since it accepts any number of files (including
  /// exactly one) in one round trip.
  Future<BulkUploadResult> uploadImages(
    List<XFile> files, {
    List<String>? tags,
  }) async {
    final json = await _client.postMultipart(
      '/api/v1/images/multiple',
      fileField: 'files',
      files: files,
      tags: (tags == null || tags.isEmpty) ? null : tags.join(','),
    );
    return BulkUploadResult.fromJson(json as Map<String, dynamic>);
  }

  List<ImageAsset> _parseList(dynamic json) =>
      (json as List<dynamic>? ?? const [])
          .map((item) => ImageAsset.fromJson(item as Map<String, dynamic>))
          .toList();
}
