import '../models/bulk_upload_result.dart';
import '../services/media_library.dart';

/// The ids of the assets in [batch] that [result] didn't report as failed.
/// The server names failures `"<filename>: <error>"`, so an asset counts as
/// failed when any entry starts with its file name; two assets sharing a
/// name both count as failed, which only costs a re-upload next time.
List<String> succeededAssetIds(
  List<ResolvedAsset> batch,
  BulkUploadResult result,
) => [
  for (final asset in batch)
    if (!result.failedFiles.any((f) => f.startsWith('${asset.file.name}: ')))
      asset.id,
];
