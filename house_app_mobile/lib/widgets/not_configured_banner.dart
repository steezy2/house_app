import 'package:flutter/material.dart';

import '../theme/app_colors.dart';

/// Shown on the Upload and Library screens when no server address has been
/// saved yet, pointing the user at the Settings tab instead of letting
/// them hit a confusing network error first.
class NotConfiguredBanner extends StatelessWidget {
  const NotConfiguredBanner({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.amber.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.amber.withValues(alpha: 0.4)),
      ),
      child: const Row(
        children: [
          Icon(Icons.warning_amber_rounded, color: AppColors.amber),
          SizedBox(width: 12),
          Expanded(
            child: Text(
              'Set your House App server address in the Settings tab before uploading.',
              style: TextStyle(color: AppColors.text),
            ),
          ),
        ],
      ),
    );
  }
}
