import 'dart:io';

import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:provider/provider.dart';

import '../../api/api.dart';
import '../../models/bulk_upload_result.dart';
import '../../services/media_library.dart';
import '../../state/settings_state.dart';
import '../../theme/app_colors.dart';
import '../../utils/batching.dart';
import '../../utils/error_message.dart';
import '../../widgets/not_configured_banner.dart';

/// The primary screen: pick photos from the gallery, optionally tag them,
/// and upload them all in one tap — the "select photos and upload" flow
/// this whole app exists for.
class UploadScreen extends StatefulWidget {
  const UploadScreen({super.key, this.mediaLibrary});

  /// Overridable for tests; defaults to a real [MediaLibrary] backed by
  /// `photo_manager`'s platform channels.
  final MediaLibrary? mediaLibrary;

  @override
  State<UploadScreen> createState() => _UploadScreenState();
}

class _UploadScreenState extends State<UploadScreen> {
  final _picker = ImagePicker();
  final _tagsController = TextEditingController();
  late final MediaLibrary _mediaLibrary;

  List<XFile> _selected = [];
  bool _uploading = false;
  BulkUploadResult? _lastResult;
  String? _error;

  bool _massUploading = false;
  int _massUploadTotal = 0;
  int _massUploadCompleted = 0;
  BulkUploadResult? _massUploadResult;
  String? _massUploadError;

  @override
  void initState() {
    super.initState();
    _mediaLibrary = widget.mediaLibrary ?? MediaLibrary();
  }

  @override
  void dispose() {
    _tagsController.dispose();
    super.dispose();
  }

  List<String> _parseTagsInput() => _tagsController.text
      .split(',')
      .map((tag) => tag.trim())
      .where((tag) => tag.isNotEmpty)
      .toList();

  Future<void> _pickPhotos() async {
    final files = await _picker.pickMultiImage(imageQuality: 90);
    if (files.isEmpty) return;
    setState(() {
      _selected = files;
      _lastResult = null;
      _error = null;
    });
  }

  Future<void> _upload() async {
    if (_selected.isEmpty) return;
    final settings = context.read<SettingsState>();
    if (!settings.isConfigured) {
      setState(
        () => _error = 'Set your server address in Settings first.',
      );
      return;
    }

    setState(() {
      _uploading = true;
      _error = null;
    });

    try {
      final result = await context.read<Api>().images.uploadImages(
        _selected,
        tags: _parseTagsInput(),
      );
      if (!mounted) return;
      setState(() {
        _lastResult = result;
        _selected = [];
        _tagsController.clear();
      });
    } catch (e) {
      if (!mounted) return;
      setState(
        () => _error = describeApiError(
          e,
          unauthorizedMessage:
              'The server rejected the API key. Check it in Settings.',
        ),
      );
    } finally {
      if (mounted) setState(() => _uploading = false);
    }
  }

  Future<void> _uploadAll() async {
    final settings = context.read<SettingsState>();
    if (!settings.isConfigured) {
      setState(
        () => _massUploadError = 'Set your server address in Settings first.',
      );
      return;
    }

    final access = await _mediaLibrary.requestAccess();
    if (!mounted) return;
    if (access == MediaAccessResult.denied) {
      setState(
        () => _massUploadError =
            'Photo library access was denied. Enable it for House App in '
            "your device's Settings app to use Upload All.",
      );
      return;
    }

    setState(() {
      _massUploadError = null;
      _massUploadResult = null;
    });

    final assets = await _mediaLibrary.getAllAssets();
    if (!mounted) return;

    if (assets.isEmpty) {
      setState(
        () => _massUploadError = 'No photos or videos found on this device.',
      );
      return;
    }

    final confirmed = await _confirmMassUpload(
      count: assets.length,
      limitedAccess: access == MediaAccessResult.limited,
    );
    if (!mounted || !confirmed) return;

    setState(() {
      _massUploading = true;
      _massUploadTotal = assets.length;
      _massUploadCompleted = 0;
    });

    final tags = _parseTagsInput();
    final batches = batchIndicesBySize(
      assets.map((a) => a.sizeBytes).toList(),
    );

    try {
      final api = context.read<Api>();
      final results = <BulkUploadResult>[];
      for (final batchIndices in batches) {
        final batchFiles = [for (final i in batchIndices) assets[i].file];
        final result = await api.images.uploadImages(batchFiles, tags: tags);
        results.add(result);
        if (!mounted) return;
        setState(() => _massUploadCompleted += batchFiles.length);
      }
      if (!mounted) return;
      setState(() => _massUploadResult = _combineResults(results));
    } catch (e) {
      if (!mounted) return;
      setState(
        () => _massUploadError = describeApiError(
          e,
          unauthorizedMessage:
              'The server rejected the API key. Check it in Settings.',
        ),
      );
    } finally {
      if (mounted) setState(() => _massUploading = false);
    }
  }

  Future<bool> _confirmMassUpload({
    required int count,
    required bool limitedAccess,
  }) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Upload everything?'),
        content: Text(
          'This uploads all $count photo${count == 1 ? '' : 's'} and '
          'video${count == 1 ? '' : 's'} ${limitedAccess ? 'you\'ve given House App access to' : 'on this device'} '
          'to your House App server. It may take a while and use a lot of '
          'data — nothing is deleted from this device.'
          '${limitedAccess ? '\n\nYou\'ve only granted access to some photos (iOS\'s "Limited Photos" mode) — this will not be everything on your device unless you allow full access in Settings.' : ''}',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Upload All'),
          ),
        ],
      ),
    );
    return confirmed ?? false;
  }

  BulkUploadResult _combineResults(List<BulkUploadResult> results) {
    return BulkUploadResult(
      totalImages: results.fold(0, (sum, r) => sum + r.totalImages),
      successfulCount: results.fold(0, (sum, r) => sum + r.successfulCount),
      failedCount: results.fold(0, (sum, r) => sum + r.failedCount),
      failedFiles: results.expand((r) => r.failedFiles).toList(),
    );
  }

  @override
  Widget build(BuildContext context) {
    final settings = context.watch<SettingsState>();

    return Scaffold(
      appBar: AppBar(title: const Text('Upload Photos')),
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.all(20),
          children: [
            if (!settings.loading && !settings.isConfigured) ...[
              const NotConfiguredBanner(),
              const SizedBox(height: 20),
            ],
            OutlinedButton.icon(
              onPressed: _uploading ? null : _pickPhotos,
              icon: const Icon(Icons.add_photo_alternate_outlined),
              label: Text(
                _selected.isEmpty
                    ? 'Select Photos'
                    : '${_selected.length} photo${_selected.length == 1 ? '' : 's'} selected — tap to change',
              ),
            ),
            if (_selected.isNotEmpty) ...[
              const SizedBox(height: 16),
              _SelectedPhotosGrid(files: _selected),
            ],
            const SizedBox(height: 16),
            TextField(
              controller: _tagsController,
              enabled: !_uploading,
              decoration: const InputDecoration(
                labelText: 'Tags (optional, comma-separated)',
                hintText: 'family, 2026',
              ),
            ),
            const SizedBox(height: 20),
            ElevatedButton(
              onPressed: (_selected.isEmpty || _uploading) ? null : _upload,
              child: _uploading
                  ? const SizedBox(
                      height: 20,
                      width: 20,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        color: Colors.black,
                      ),
                    )
                  : Text(
                      _selected.isEmpty
                          ? 'Upload'
                          : 'Upload ${_selected.length} photo${_selected.length == 1 ? '' : 's'}',
                    ),
            ),
            if (_error != null) ...[
              const SizedBox(height: 16),
              Text(_error!, style: const TextStyle(color: AppColors.red)),
            ],
            if (_lastResult != null) ...[
              const SizedBox(height: 20),
              _UploadResultCard(result: _lastResult!),
            ],
            const SizedBox(height: 28),
            const Divider(),
            const SizedBox(height: 16),
            Text(
              'Or grab everything',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 4),
            const Text(
              'Uploads every photo and video already on this device, in '
              'batches. Requires full photo library access — a broader '
              'permission than "Select Photos" above needs.',
              style: TextStyle(color: AppColors.textSecondary),
            ),
            const SizedBox(height: 12),
            ElevatedButton.icon(
              onPressed: _massUploading ? null : _uploadAll,
              icon: const Icon(Icons.cloud_upload_outlined),
              label: Text(
                _massUploading
                    ? 'Uploading...'
                    : 'Upload All Photos & Videos',
              ),
            ),
            if (_massUploading) ...[
              const SizedBox(height: 12),
              LinearProgressIndicator(
                value: _massUploadTotal == 0
                    ? null
                    : _massUploadCompleted / _massUploadTotal,
              ),
              const SizedBox(height: 4),
              Text(
                '$_massUploadCompleted / $_massUploadTotal uploaded',
                style: const TextStyle(color: AppColors.textSecondary),
              ),
            ],
            if (_massUploadError != null) ...[
              const SizedBox(height: 12),
              Text(
                _massUploadError!,
                style: const TextStyle(color: AppColors.red),
              ),
            ],
            if (_massUploadResult != null) ...[
              const SizedBox(height: 12),
              _UploadResultCard(result: _massUploadResult!),
            ],
          ],
        ),
      ),
    );
  }
}

class _SelectedPhotosGrid extends StatelessWidget {
  const _SelectedPhotosGrid({required this.files});

  final List<XFile> files;

  @override
  Widget build(BuildContext context) {
    return GridView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: files.length,
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 4,
        crossAxisSpacing: 8,
        mainAxisSpacing: 8,
      ),
      itemBuilder: (context, index) => ClipRRect(
        borderRadius: BorderRadius.circular(8),
        child: Image.file(File(files[index].path), fit: BoxFit.cover),
      ),
    );
  }
}

class _UploadResultCard extends StatelessWidget {
  const _UploadResultCard({required this.result});

  final BulkUploadResult result;

  @override
  Widget build(BuildContext context) {
    final color = result.isFullSuccess ? AppColors.green : AppColors.amber;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  result.isFullSuccess
                      ? Icons.check_circle_outline
                      : Icons.error_outline,
                  color: color,
                ),
                const SizedBox(width: 8),
                Text(
                  '${result.successfulCount} of ${result.totalImages} uploaded',
                  style: Theme.of(
                    context,
                  ).textTheme.titleMedium?.copyWith(color: color),
                ),
              ],
            ),
            if (result.failedFiles.isNotEmpty) ...[
              const SizedBox(height: 8),
              for (final failure in result.failedFiles)
                Text(
                  failure,
                  style: const TextStyle(color: AppColors.textSecondary),
                ),
            ],
          ],
        ),
      ),
    );
  }
}
