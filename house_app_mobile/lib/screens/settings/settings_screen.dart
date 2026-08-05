import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../../api/api.dart';
import '../../state/settings_state.dart';
import '../../theme/app_colors.dart';
import '../../utils/error_message.dart';

/// Where the server address and API key (see house_app_server's `API_KEY`
/// env var / `X-API-Key` header) are entered once, then reused by every
/// other screen via [SettingsState].
class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  final _baseUrlController = TextEditingController();
  final _apiKeyController = TextEditingController();

  bool _obscureKey = true;
  bool _testing = false;
  bool? _testSucceeded;
  String? _testMessage;
  bool _fieldsHydrated = false;

  @override
  void dispose() {
    _baseUrlController.dispose();
    _apiKeyController.dispose();
    super.dispose();
  }

  Future<void> _save({bool silent = false}) async {
    await context.read<SettingsState>().save(
      baseUrl: _baseUrlController.text,
      apiKey: _apiKeyController.text.trim(),
    );
    if (!mounted || silent) return;
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(const SnackBar(content: Text('Settings saved')));
  }

  Future<void> _testConnection() async {
    // Read before the first await, not after — context shouldn't be used
    // across an async gap.
    final api = context.read<Api>();

    // Save first so the test exercises whatever is currently typed, not
    // whatever was last saved.
    await _save(silent: true);
    if (!mounted) return;

    setState(() {
      _testing = true;
      _testMessage = null;
      _testSucceeded = null;
    });

    try {
      await api.images.getImages();
      if (!mounted) return;
      setState(() {
        _testSucceeded = true;
        _testMessage = 'Connected successfully.';
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _testSucceeded = false;
        _testMessage = describeApiError(
          e,
          unauthorizedMessage: 'Connected, but the server rejected the API key.',
        );
      });
    } finally {
      if (mounted) setState(() => _testing = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final settings = context.watch<SettingsState>();

    if (!_fieldsHydrated && !settings.loading) {
      _baseUrlController.text = settings.baseUrl ?? '';
      _apiKeyController.text = settings.apiKey ?? '';
      _fieldsHydrated = true;
    }

    return Scaffold(
      appBar: AppBar(title: const Text('Settings')),
      body: settings.loading
          ? const Center(child: CircularProgressIndicator())
          : SafeArea(
              child: ListView(
                padding: const EdgeInsets.all(20),
                children: [
                  Text(
                    'Server address',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 4),
                  const Text(
                    'Your House App server on the home network, e.g. '
                    '192.168.1.23:8080. "http://" is assumed if you leave '
                    'it off.',
                    style: TextStyle(color: AppColors.textSecondary),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: _baseUrlController,
                    keyboardType: TextInputType.url,
                    autocorrect: false,
                    decoration: const InputDecoration(
                      labelText: 'Server address',
                      hintText: '192.168.1.23:8080',
                    ),
                  ),
                  const SizedBox(height: 24),
                  Text(
                    'API key',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 4),
                  const Text(
                    'Must match the API_KEY set on the server. Leave blank '
                    'if the server has none configured.',
                    style: TextStyle(color: AppColors.textSecondary),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: _apiKeyController,
                    obscureText: _obscureKey,
                    autocorrect: false,
                    decoration: InputDecoration(
                      labelText: 'API key',
                      suffixIcon: IconButton(
                        icon: Icon(
                          _obscureKey
                              ? Icons.visibility_outlined
                              : Icons.visibility_off_outlined,
                        ),
                        onPressed: () =>
                            setState(() => _obscureKey = !_obscureKey),
                      ),
                    ),
                  ),
                  const SizedBox(height: 28),
                  ElevatedButton(
                    onPressed: () => _save(),
                    child: const Text('Save'),
                  ),
                  const SizedBox(height: 12),
                  OutlinedButton(
                    onPressed: _testing ? null : _testConnection,
                    child: _testing
                        ? const SizedBox(
                            height: 18,
                            width: 18,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('Test Connection'),
                  ),
                  if (_testMessage != null) ...[
                    const SizedBox(height: 16),
                    Text(
                      _testMessage!,
                      style: TextStyle(
                        color: _testSucceeded == true
                            ? AppColors.green
                            : AppColors.red,
                      ),
                    ),
                  ],
                ],
              ),
            ),
    );
  }
}
