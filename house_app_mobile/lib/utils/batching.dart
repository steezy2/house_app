/// Splits indices `[0, sizesInBytes.length)` into batches, each capped at
/// [maxCount] items and [maxBytesPerBatch] combined bytes — whichever
/// limit is hit first — so the "Upload All" flow never sends an unbounded
/// number of files or an unbounded amount of data in a single HTTP
/// request. A single item larger than [maxBytesPerBatch] still gets its
/// own one-item batch rather than being dropped or blocking everything
/// queued after it.
///
/// Returns indices rather than the items themselves so this stays a pure
/// function with no dependency on `photo_manager`/`image_picker` types,
/// which keeps it trivially unit-testable.
List<List<int>> batchIndicesBySize(
  List<int> sizesInBytes, {
  int maxCount = 20,
  int maxBytesPerBatch = 150 * 1024 * 1024,
}) {
  final batches = <List<int>>[];
  var current = <int>[];
  var currentBytes = 0;

  for (var i = 0; i < sizesInBytes.length; i++) {
    final size = sizesInBytes[i];
    final exceedsCount = current.length >= maxCount;
    final exceedsBytes =
        current.isNotEmpty && currentBytes + size > maxBytesPerBatch;

    if (exceedsCount || exceedsBytes) {
      batches.add(current);
      current = [];
      currentBytes = 0;
    }

    current.add(i);
    currentBytes += size;
  }

  if (current.isNotEmpty) {
    batches.add(current);
  }

  return batches;
}
