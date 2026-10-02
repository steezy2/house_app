import 'package:image_picker/image_picker.dart';
import 'package:photo_manager/photo_manager.dart';

/// A photo/video asset resolved to a real file on disk, ready to upload.
class ResolvedAsset {
  const ResolvedAsset({
    required this.id,
    required this.file,
    required this.sizeBytes,
  });

  /// `photo_manager`'s stable per-device asset id.
  final String id;
  final XFile file;
  final int sizeBytes;
}

/// The outcome of requesting full photo/video library access.
enum MediaAccessResult {
  /// Full access — every asset on the device is enumerable.
  full,

  /// iOS's "Limited Photos" mode: the user granted access to only the
  /// assets they explicitly picked in the system permission sheet, not
  /// everything. [MediaLibrary.getAllAssets] only sees that subset — this
  /// is surfaced separately from [full] so callers can be honest that an
  /// "upload everything" pass wasn't actually everything.
  limited,

  /// Access denied outright.
  denied,
}

/// Wraps `photo_manager`'s full-camera-roll access behind a small,
/// upload-focused API. Unlike `image_picker`'s `pickMultiImage` (a
/// one-shot picker widget the user selects individual items from), this
/// enumerates every photo and video already on the device — the "grab
/// everything" capability the mass-upload button needs — which requires
/// the broader OS "photo library access" permission rather than a
/// picker-scoped one that needs no permission at all.
class MediaLibrary {
  /// Requests full photo/video library access. See [MediaAccessResult]
  /// for what each outcome means.
  Future<MediaAccessResult> requestAccess() async {
    final state = await PhotoManager.requestPermissionExtend();
    if (state.isAuth) return MediaAccessResult.full;
    if (state.hasAccess) return MediaAccessResult.limited;
    return MediaAccessResult.denied;
  }

  /// Enumerates every image and video asset accessible to this app
  /// (everything, if [requestAccess] returned [MediaAccessResult.full];
  /// only the user-picked subset if it returned
  /// [MediaAccessResult.limited]), resolved to real files with their
  /// sizes. Newest first, matching `photo_manager`'s default ordering.
  /// Assets whose id is in [exclude] are skipped before their file is
  /// resolved, which on iOS can mean downloading it from iCloud.
  Future<List<ResolvedAsset>> getAllAssets({
    Set<String> exclude = const {},
  }) async {
    final albums = await PhotoManager.getAssetPathList(
      type: RequestType.common,
      onlyAll: true,
    );
    if (albums.isEmpty) return const [];

    // onlyAll: true guarantees exactly one synthetic "all assets" album
    // instead of one per device album, so nothing is double-counted.
    final album = albums.first;
    final count = await album.assetCountAsync;
    final assets = await album.getAssetListRange(start: 0, end: count);

    final resolved = <ResolvedAsset>[];
    for (final asset in assets) {
      if (exclude.contains(asset.id)) continue;
      final file = await asset.originFile;
      // e.g. an iCloud-only asset that failed to download
      if (file == null) continue;
      final size = await file.length();
      resolved.add(
        ResolvedAsset(id: asset.id, file: XFile(file.path), sizeBytes: size),
      );
    }
    return resolved;
  }
}
