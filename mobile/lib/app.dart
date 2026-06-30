import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'features/auth/providers/auth_provider.dart';
import 'features/auth/providers/biometric_provider.dart';
import 'features/auth/screens/login_screen.dart';
import 'features/auth/screens/biometric_lock_screen.dart';
import 'features/home/screens/home_screen.dart';

class QuiverScoreApp extends ConsumerWidget {
  const QuiverScoreApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);
    final isLocked = ref.watch(biometricLockStateProvider);

    return MaterialApp(
      title: 'QuiverScore',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF2E7D32), // archery green
          brightness: Brightness.light,
        ),
        useMaterial3: true,
      ),
      darkTheme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF2E7D32),
          brightness: Brightness.dark,
        ),
        useMaterial3: true,
      ),
      home: !authState.isAuthenticated
          ? const LoginScreen()
          : isLocked
              ? const BiometricLockScreen()
              : const HomeScreen(),
    );
  }
}
