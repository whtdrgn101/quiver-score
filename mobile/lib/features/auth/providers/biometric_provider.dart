import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/storage/secure_storage.dart';
import '../../../core/biometric/biometric_service.dart';

final biometricLockSettingProvider =
    StateNotifierProvider<BiometricLockSettingNotifier, bool>((ref) {
  return BiometricLockSettingNotifier(ref.watch(secureStorageProvider));
});

class BiometricLockSettingNotifier extends StateNotifier<bool> {
  final SecureStorage _storage;

  BiometricLockSettingNotifier(this._storage) : super(false) {
    _loadSetting();
  }

  Future<void> _loadSetting() async {
    state = await _storage.isBiometricLockEnabled();
  }

  Future<void> toggle(bool enabled) async {
    await _storage.setBiometricLockEnabled(enabled);
    state = enabled;
  }
}

final biometricLockStateProvider =
    StateNotifierProvider<BiometricLockStateNotifier, bool>((ref) {
  return BiometricLockStateNotifier(ref);
});

class BiometricLockStateNotifier extends StateNotifier<bool> {
  BiometricLockStateNotifier(Ref ref) : super(false) {
    _init(ref);
  }

  Future<void> _init(Ref ref) async {
    final storage = ref.read(secureStorageProvider);
    final enabled = await storage.isBiometricLockEnabled();
    state = enabled;
  }

  void unlock() {
    state = false;
  }

  void lock() {
    state = true;
  }
}

final isBiometricsSupportedProvider = FutureProvider<bool>((ref) async {
  final service = ref.watch(biometricServiceProvider);
  return await service.canAuthenticate();
});
