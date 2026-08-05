import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';

import '../../api/api.dart';
import '../../models/image_asset.dart';
import '../../state/settings_state.dart';
import '../../theme/app_colors.dart';
import '../../widgets/not_configured_banner.dart';

/// Lists everything uploaded so far, newest first. The server has no
/// endpoint to fetch raw image bytes (see house_app CLAUDE.md), so this is
/// metadata only — filename, tags, category, and whether the background
/// processor has organized it yet — not a photo grid.
class LibraryScreen extends StatefulWidget {
  const LibraryScreen({super.key});

  @override
  State<LibraryScreen> createState() => _LibraryScreenState();
}

class _LibraryScreenState extends State<LibraryScreen> {
  Future<List<ImageAsset>>? _future;

  Future<List<ImageAsset>> _load() => context.read<Api>().images.getImages();

  Future<void> _refresh() async {
    final future = _load();
    setState(() => _future = future);
    await future;
  }

  @override
  Widget build(BuildContext context) {
    final settings = context.watch<SettingsState>();

    if (settings.loading) {
      return const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
    }

    if (!settings.isConfigured) {
      return Scaffold(
        appBar: AppBar(title: const Text('Library')),
        body: const Padding(
          padding: EdgeInsets.all(20),
          child: NotConfiguredBanner(),
        ),
      );
    }

    _future ??= _load();

    return Scaffold(
      appBar: AppBar(title: const Text('Library')),
      body: RefreshIndicator(
        onRefresh: _refresh,
        child: FutureBuilder<List<ImageAsset>>(
          future: _future,
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return const Center(child: CircularProgressIndicator());
            }
            if (snapshot.hasError) {
              return _messageList('Failed to load: ${snapshot.error}');
            }
            final images = [...?snapshot.data]
              ..sort((a, b) => b.createdAt.compareTo(a.createdAt));
            if (images.isEmpty) {
              return _messageList('No photos uploaded yet.');
            }
            return ListView.separated(
              itemCount: images.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (context, index) => _ImageTile(image: images[index]),
            );
          },
        ),
      ),
    );
  }

  /// Wraps a single centered message in a scrollable list so
  /// [RefreshIndicator]'s pull-to-refresh gesture still works on an empty
  /// or errored screen.
  Widget _messageList(String message) => ListView(
    children: [
      const SizedBox(height: 120),
      Center(
        child: Text(message, style: const TextStyle(color: AppColors.textSecondary)),
      ),
    ],
  );
}

class _ImageTile extends StatelessWidget {
  const _ImageTile({required this.image});

  final ImageAsset image;

  @override
  Widget build(BuildContext context) {
    final dateFormat = DateFormat.yMMMd().add_jm();
    final subtitleParts = [
      if (image.category != null) image.category!,
      dateFormat.format(image.createdAt.toLocal()),
      if (image.tags.isNotEmpty) image.tags.join(', '),
    ];

    return ListTile(
      leading: CircleAvatar(
        backgroundColor: AppColors.backgroundElevated2,
        child: Icon(
          image.isProcessed ? Icons.folder_outlined : Icons.hourglass_top,
          color: AppColors.textSecondary,
          size: 20,
        ),
      ),
      title: Text(
        image.filename,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
      ),
      subtitle: Text(subtitleParts.join(' · ')),
    );
  }
}
