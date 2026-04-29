import 'package:dio/dio.dart';
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
      syncService.pullRoundTemplates();
    }
  }

  Future<void> login({
    required String username,
    required String password,
  }) async {
    state = const AuthState.loading();
    try {
      final response = await api.dio.post('/api/v1/auth/login', data: {
        'username': username,
        'password': password,
      });

      await _handleAuthResponse(response.data);
    } catch (e) {
      state = AuthState.error(_extractError(e));
    }
  }

  Future<void> register({
    required String email,
    required String username,
    required String password,
    String? displayName,
  }) async {
    state = const AuthState.loading();
    try {
      final response = await api.dio.post('/api/v1/auth/register', data: {
        'email': email,
        'username': username,
        'password': password,
        if (displayName != null && displayName.isNotEmpty)
          'display_name': displayName,
      });

      await _handleAuthResponse(response.data);
    } catch (e) {
      state = AuthState.error(_extractError(e));
    }
  }

  Future<void> _handleAuthResponse(dynamic data) async {
    final accessToken = data['access_token'] as String;
    final refreshToken = data['refresh_token'] as String;

    await storage.saveTokens(
      accessToken: accessToken,
      refreshToken: refreshToken,
    );

    state = const AuthState.authenticated();

    syncService.pullRoundTemplates();
    syncService.syncPendingItems();
  }

  Future<void> logout() async {
    await storage.clearTokens();
    state = const AuthState.initial();
  }

  String _extractError(dynamic e) {
    if (e is DioException && e.response?.data is Map) {
      final data = e.response!.data as Map;
      final serverMsg = data['error'] as String?;
      if (serverMsg != null && serverMsg.isNotEmpty) return serverMsg;
    }
    if (e is DioException) {
      if (e.type == DioExceptionType.connectionError ||
          e.type == DioExceptionType.connectionTimeout) {
        return 'No internet connection. Please try again.';
      }
    }
    return 'Something went wrong. Please try again.';
  }
}
