/// Shared JSON-decoding helpers used by every model's `fromJson`.
library;

/// Go's `encoding/json` `omitempty` does not omit zero-valued (non-pointer)
/// `time.Time` structs, so `processedAt` is always present as a string on
/// the wire, just possibly Go's zero-time sentinel
/// ("0001-01-01T00:00:00Z") when the background processor hasn't touched
/// the image yet. Treat that as null.
DateTime? parseNullableDate(dynamic value) {
  if (value == null) return null;
  final str = value as String;
  if (str.isEmpty) return null;
  final date = DateTime.tryParse(str);
  if (date == null || date.year <= 1) return null;
  return date;
}

DateTime parseDate(dynamic value) =>
    parseNullableDate(value) ?? DateTime.fromMillisecondsSinceEpoch(0);

String? asNullableString(dynamic value) {
  final str = value as String?;
  return (str == null || str.isEmpty) ? null : str;
}

List<String> asStringList(dynamic value) =>
    (value as List<dynamic>?)?.map((item) => item as String).toList() ??
    const [];
