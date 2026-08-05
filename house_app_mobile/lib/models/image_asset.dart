import 'json_utils.dart';

/// Mirrors house_app_server/models/models.go's `Image`.
class ImageAsset {
  const ImageAsset({
    required this.id,
    required this.filename,
    required this.originalPath,
    required this.storagePath,
    required this.size,
    required this.contentType,
    required this.tags,
    required this.category,
    required this.createdAt,
    required this.processedAt,
  });

  final String id;
  final String filename;
  final String originalPath;
  final String storagePath;
  final int size;
  final String contentType;
  final List<String> tags;
  final String? category;
  final DateTime createdAt;
  final DateTime? processedAt;

  /// True once the background processor on the server has moved this file
  /// out of the upload landing zone into its organized, dated location.
  /// There's no endpoint to fetch the actual image bytes, so the app can
  /// only ever show this metadata, not a thumbnail.
  bool get isProcessed => processedAt != null;

  factory ImageAsset.fromJson(Map<String, dynamic> json) => ImageAsset(
    id: json['id'] as String? ?? '',
    filename: json['filename'] as String? ?? '',
    originalPath: json['originalPath'] as String? ?? '',
    storagePath: json['storagePath'] as String? ?? '',
    size: (json['size'] as num?)?.toInt() ?? 0,
    contentType: json['contentType'] as String? ?? '',
    tags: asStringList(json['tags']),
    category: asNullableString(json['category']),
    createdAt: parseDate(json['createdAt']),
    processedAt: parseNullableDate(json['processedAt']),
  );
}
