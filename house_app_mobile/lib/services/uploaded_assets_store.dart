import 'dart:io';

import 'package:path_provider/path_provider.dart';

/// Remembers which device assets (by `photo_manager` asset id) already
/// reached the server, so "Upload All" only sends new photos and videos
/// and an interrupted run picks up where it stopped. Stored as one id per
/// line in the app's support directory: a plain file rather than secure
/// storage or preferences, since a whole camera roll is tens of thousands
/// of ids and none of them is a secret.
class UploadedAssetsStore {
  UploadedAssetsStore({Future<Directory> Function()? directory})
    : _directory = directory ?? getApplicationSupportDirectory;

  final Future<Directory> Function() _directory;
  static const _fileName = 'uploaded_asset_ids.txt';

  Future<File> _file() async =>
      File('${(await _directory()).path}${Platform.pathSeparator}$_fileName');

  /// The ids recorded so far; empty before the first upload.
  Future<Set<String>> load() async {
    final file = await _file();
    if (!await file.exists()) return {};
    return (await file.readAsLines()).where((id) => id.isNotEmpty).toSet();
  }

  /// Records [ids] as uploaded.
  Future<void> addAll(Iterable<String> ids) async {
    if (ids.isEmpty) return;
    final file = await _file();
    await file.writeAsString(
      '${ids.join('\n')}\n',
      mode: FileMode.append,
      flush: true,
    );
  }
}
