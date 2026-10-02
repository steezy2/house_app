import 'package:house_app_mobile/services/uploaded_assets_store.dart';

/// In-memory [UploadedAssetsStore] for widget tests, which can't do real
/// file I/O (see the note in `test/screens/upload_screen_test.dart`).
class FakeUploadedAssetsStore extends UploadedAssetsStore {
  FakeUploadedAssetsStore([Set<String>? ids]) : ids = ids ?? {};

  final Set<String> ids;

  @override
  Future<Set<String>> load() async => {...ids};

  @override
  Future<void> addAll(Iterable<String> newIds) async => ids.addAll(newIds);
}
