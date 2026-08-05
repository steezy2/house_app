import '../api/api_client.dart';

/// Extracts a user-facing message from a caught upload/connection error —
/// shared by the Upload and Settings screens, which otherwise each ran the
/// same [ApiConfigException]/[ApiException]/generic classification.
///
/// [unauthorizedMessage] lets a call site customize the 401 case, since the
/// right phrasing differs by context (e.g. "can't upload yet" vs. "test
/// connection failed"); it falls back to a generic message if omitted.
String describeApiError(Object error, {String? unauthorizedMessage}) {
  if (error is ApiConfigException) return error.message;
  if (error is ApiException) {
    if (error.statusCode == 401) {
      return unauthorizedMessage ?? 'The server rejected the API key.';
    }
    return error.message;
  }
  return 'Could not reach the server: $error';
}
