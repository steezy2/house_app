import 'dart:io';

import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:provider/provider.dart';

import '../../api/api.dart';
import '../../api/api_client.dart';
import '../../models/bulk_upload_result.dart';
import '../../state/settings_state.dart';
import '../../theme/app_colors.dart';
import '../../widgets/not_configured_banner.dart';

/// The primary screen: pick photos from the gallery, optionally tag them,
/// and upload them all in one tap — the "select photos and upload" flow
/// this whole app exists for.
class UploadScreen extends StatefulWidget {
  const UploadScreen({super.key});

  @override
  State<UploadScreen> createState() => _UploadScreenState();
}

class _UploadScreenState extends State<UploadScreen> {
  final _picker = ImagePicker();
  final _tagsController = TextEditingController();

  List<XFile> _selected = [];
  bool _uploading = false;
  BulkUploadResult? _lastResult;
  String? _error;

  @override
  void dispose() {
    _tagsController.dispose();
    super.dispose();
  }

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

    final tags = _tagsController.text
        .split(',')
        .map((tag) => tag.trim())
        .where((tag) => tag.isNotEmpty)
        .toList();

    try {
      final result = await context.read<Api>().images.uploadImages(
        _selected,
        tags: tags,
      );
      if (!mounted) return;
      setState(() {
        _lastResult = result;
        _selected = [];
        _tagsController.clear();
      });
    } on ApiConfigException catch (e) {
      if (!mounted) return;
      setState(() => _error = e.message);
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(
        () => _error = e.statusCode == 401
            ? 'The server rejected the API key. Check it in Settings.'
            : e.message,
      );
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = 'Could not reach the server: $e');
    } finally {
      if (mounted) setState(() => _uploading = false);
    }
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
