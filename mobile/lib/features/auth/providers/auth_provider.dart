import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';
import '../../../core/storage/secure_storage.dart';
import '../../../core/sync/sync_service.dart';
import '../models/auth_state.dart';

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(
    api: ref.watch(apiClientProvider),
    storage: ref.watch(secureStorageProvider),
    syncService: ref.watch(syncServiceProvider),
  );
});

class AuthNotifier extends StateNotifier<AuthState> {
  final ApiClient api;
  final SecureStorage storage;
  final SyncService syncService;

  AuthNotifier({
    required this.api,
    required this.storage,
    required this.syncService,
  }) : super(const AuthState.initial()) {
    _tryRestoreSession();
  }

  Future<void> _tryRestoreSession() async {
    final token = await storage.getAccessToken();
    if (token != null) {
      state = const AuthState.authenticated();
      // Pull latest round templates on session restore
      syncService.pullRoundTemplates();
    }
  }

  Future<void> login({
    required String email,
    required String password,
  }) async {
    state = const AuthState.loading();
    try {
      final response = await api.dio.post('/api/v1/auth/login', data: {
        'email': email,
        'password': password,
      });

      final accessToken = response.data['access_token'] as String;
      final refreshToken = response.data['refresh_token'] as String;

      await storage.saveTokens(
        accessToken: accessToken,
        refreshToken: refreshToken,
      );

      state = const AuthState.authenticated();

      // Sync round templates after login
      syncService.pullRoundTemplates();
      syncService.syncPendingItems();
    } catch (e) {
      state = AuthState.error(_extractError(e));
    }
  }

  Future<void> logout() async {
    await storage.clearTokens();
    state = const AuthState.initial();
  }

  String _extractError(dynamic e) {
    if (e is Exception) {
      return e.toString().replaceAll('Exception: ', '');
    }
    return 'Login failed. Please try again.';
  }
}
