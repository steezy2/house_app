/// Mirrors house_app_server/models/models.go's `BulkUploadResponse`,
/// returned by `POST /api/v1/images/multiple`.
class BulkUploadResult {
  const BulkUploadResult({
    required this.totalImages,
    required this.successfulCount,
    required this.failedCount,
    required this.failedFiles,
  });

  final int totalImages;
  final int successfulCount;
  final int failedCount;

  /// `"<filename>: <error message>"` entries, one per failed file.
  final List<String> failedFiles;

  bool get isFullSuccess => totalImages > 0 && failedCount == 0;

  factory BulkUploadResult.fromJson(Map<String, dynamic> json) =>
      BulkUploadResult(
        totalImages: (json['totalImages'] as num?)?.toInt() ?? 0,
        successfulCount: (json['successfulCount'] as num?)?.toInt() ?? 0,
        failedCount: (json['failedCount'] as num?)?.toInt() ?? 0,
        failedFiles:
            (json['failedFiles'] as List<dynamic>?)
                ?.map((item) => item as String)
                .toList() ??
            const [],
      );
}
