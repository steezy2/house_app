import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import 'api/api.dart';
import 'state/settings_state.dart';
import 'state/settings_storage.dart';
import 'theme/app_theme.dart';
import 'widgets/bottom_nav_shell.dart';

void main() {
  runApp(const HouseAppMobileApp());
}

class HouseAppMobileApp extends StatelessWidget {
  const HouseAppMobileApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        Provider(create: (_) => SettingsStorage()),
        Provider(
          create: (context) =>
              Api(settingsStorage: context.read<SettingsStorage>()),
        ),
        ChangeNotifierProvider(
          create: (context) =>
              SettingsState(context.read<SettingsStorage>())..hydrate(),
        ),
      ],
      child: MaterialApp(
        title: 'House App',
        theme: AppTheme.dark,
        home: const BottomNavShell(),
      ),
    );
  }
}
