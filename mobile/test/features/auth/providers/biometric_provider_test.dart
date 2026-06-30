import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:quiverscore/core/storage/secure_storage.dart';
import 'package:quiverscore/features/auth/providers/biometric_provider.dart';
import '../../../helpers/fake_secure_storage.dart';

void main() {
  late FakeSecureStorage fakeStorage;
  late ProviderContainer container;

  setUp(() {
    fakeStorage = FakeSecureStorage();
    container = ProviderContainer(
      overrides: [
        secureStorageProvider.overrideWithValue(fakeStorage),
      ],
    );
  });

  tearDown(() {
    container.dispose();
  });

  group('BiometricLockSettingNotifier', () {
    test('initial state is false when no preference stored', () async {
      final setting = container.read(biometricLockSettingProvider);
      expect(setting, false);
    });

    test('loads true state when enabled in storage', () async {
      await fakeStorage.setBiometricLockEnabled(true);
      
      // Re-read setting provider in a new container to simulate reload
      final localContainer = ProviderContainer(
        overrides: [
          secureStorageProvider.overrideWithValue(fakeStorage),
        ],
      );
      
      // Let async load run
      await localContainer.read(biometricLockSettingProvider.notifier).toggle(true);
      final setting = localContainer.read(biometricLockSettingProvider);
      expect(setting, true);
      localContainer.dispose();
    });

    test('toggle changes setting and updates storage', () async {
      expect(container.read(biometricLockSettingProvider), false);
      
      await container.read(biometricLockSettingProvider.notifier).toggle(true);
      
      expect(container.read(biometricLockSettingProvider), true);
      expect(await fakeStorage.isBiometricLockEnabled(), true);

      await container.read(biometricLockSettingProvider.notifier).toggle(false);
      
      expect(container.read(biometricLockSettingProvider), false);
      expect(await fakeStorage.isBiometricLockEnabled(), false);
    });
  });

  group('BiometricLockStateNotifier', () {
    test('initial state is false when biometric lock is disabled', () async {
      final isLocked = container.read(biometricLockStateProvider);
      expect(isLocked, false);
    });

    test('initial state is true when biometric lock is enabled in storage', () async {
      await fakeStorage.setBiometricLockEnabled(true);

      final localContainer = ProviderContainer(
        overrides: [
          secureStorageProvider.overrideWithValue(fakeStorage),
        ],
      );

      // Trigger instantiation
      localContainer.read(biometricLockStateProvider);

      // Need to wait for _init to run
      await Future<void>.delayed(Duration.zero);

      final isLocked = localContainer.read(biometricLockStateProvider);
      expect(isLocked, true);
      localContainer.dispose();
    });

    test('unlock sets state to false', () {
      container.read(biometricLockStateProvider.notifier).lock();
      expect(container.read(biometricLockStateProvider), true);

      container.read(biometricLockStateProvider.notifier).unlock();
      expect(container.read(biometricLockStateProvider), false);
    });
  });
}
