import 'dart:developer' as dev;

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/legacy.dart';

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
    api.onAuthExpired = _onAuthExpired;
    _tryRestoreSession();
  }

  void _onAuthExpired() {
    storage.clearTokens();
    state = const AuthState.initial();
  }

  Future<void> _tryRestoreSession() async {
    final accessToken = await storage.getAccessToken();
    final refreshToken = await storage.getRefreshToken();
    if (accessToken != null || refreshToken != null) {
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
    dev.log('Auth error: $e', name: 'AuthNotifier');
    if (e is DioException) {
      dev.log('DioException type: ${e.type}', name: 'AuthNotifier');
      dev.log('DioException status: ${e.response?.statusCode}', name: 'AuthNotifier');
      dev.log('DioException data: ${e.response?.data}', name: 'AuthNotifier');
      dev.log('DioException message: ${e.message}', name: 'AuthNotifier');

      if (e.response?.data is Map) {
        final data = e.response!.data as Map;
        final serverMsg = data['error'] as String?;
        if (serverMsg != null && serverMsg.isNotEmpty) return serverMsg;
      }
      // Firebase Hosting may return HTML error pages
      if (e.response?.data is String) {
        final body = e.response!.data as String;
        if (body.contains('Not Found')) return 'Service not available. Please try again later.';
      }
      if (e.type == DioExceptionType.connectionError ||
          e.type == DioExceptionType.connectionTimeout) {
        return 'No internet connection. Please try again.';
      }
      if (e.response?.statusCode != null) {
        return 'Request failed (${e.response!.statusCode}). ${e.message ?? ''}';
      }
      return 'Connection error: ${e.type.name}. ${e.message ?? ''}';
    }
    return 'Unexpected error: $e';
  }
}
