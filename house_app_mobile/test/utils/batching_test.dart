import 'package:flutter_test/flutter_test.dart';
import 'package:house_app_mobile/utils/batching.dart';

void main() {
  group('batchIndicesBySize', () {
    test('returns nothing for an empty input', () {
      expect(batchIndicesBySize([]), isEmpty);
    });

    test('puts everything in one batch when well under both limits', () {
      final batches = batchIndicesBySize([100, 200, 300]);
      expect(batches, [
        [0, 1, 2],
      ]);
    });

    test('splits once the item-count cap is hit', () {
      final sizes = List.filled(25, 1024); // 25 tiny files, cap is 20
      final batches = batchIndicesBySize(sizes, maxCount: 20);

      expect(batches.length, 2);
      expect(batches[0].length, 20);
      expect(batches[1].length, 5);
    });

    test('splits once the byte-size cap is hit', () {
      const maxBytes = 150;
      // 100 + 100 exceeds 150, so the second item starts a new batch.
      final batches = batchIndicesBySize(
        [100, 100, 100],
        maxBytesPerBatch: maxBytes,
      );

      expect(batches, [
        [0],
        [1],
        [2],
      ]);
    });

    test('gives an oversized single item its own batch rather than dropping it', () {
      final batches = batchIndicesBySize(
        [10, 500, 10],
        maxBytesPerBatch: 150,
      );

      expect(batches, [
        [0],
        [1],
        [2],
      ]);
    });

    test('resumes filling normally after an oversized item', () {
      final batches = batchIndicesBySize(
        [10, 500, 10, 10],
        maxBytesPerBatch: 150,
      );

      expect(batches, [
        [0],
        [1],
        [2, 3],
      ]);
    });

    test('respects both caps together', () {
      // maxCount=2 forces a split after every 2 items even though the
      // byte cap alone wouldn't require it.
      final batches = batchIndicesBySize(
        [10, 10, 10, 10, 10],
        maxCount: 2,
        maxBytesPerBatch: 1000,
      );

      expect(batches, [
        [0, 1],
        [2, 3],
        [4],
      ]);
    });
  });
}
