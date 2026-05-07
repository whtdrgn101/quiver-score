import 'dart:ui';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:meta/meta.dart';

import '../storage/secure_storage.dart';

enum RefreshResult { refreshed, rejected, networkError }

final apiClientProvider = Provider<ApiClient>((ref) {
  final storage = ref.watch(secureStorageProvider);
  return ApiClient(storage: storage);
});

class ApiClient {
  // Production: Firebase Hosting proxies /api/** to Cloud Run
  static const String _defaultBaseUrl = 'https://quiverscore.com';

  final SecureStorage storage;
  late final Dio dio;
  final Dio Function() _refreshDioFactory;

  VoidCallback? onAuthExpired;

  ApiClient({
    required this.storage,
    String? baseUrl,
    @visibleForTesting Dio Function()? refreshDioFactory,
  }) : _refreshDioFactory = refreshDioFactory ?? Dio.new {
    dio = Dio(BaseOptions(
      baseUrl: baseUrl ?? _defaultBaseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 15),
      headers: {'Content-Type': 'application/json'},
    ));

    dio.interceptors.add(
        _AuthInterceptor(storage, dio, this, _refreshDioFactory));
  }
}

class _AuthInterceptor extends Interceptor {
  final SecureStorage _storage;
  final Dio _dio;
  final ApiClient _client;
  final Dio Function() _refreshDioFactory;

  _AuthInterceptor(
      this._storage, this._dio, this._client, this._refreshDioFactory);

  @override
  void onRequest(
      RequestOptions options, RequestInterceptorHandler handler) async {
    final token = await _storage.getAccessToken();
    if (token != null) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    if (err.response?.statusCode == 401) {
      final result = await _tryRefresh();
      switch (result) {
        case RefreshResult.refreshed:
          final token = await _storage.getAccessToken();
          err.requestOptions.headers['Authorization'] = 'Bearer $token';
          try {
            final response = await _dio.fetch(err.requestOptions);
            return handler.resolve(response);
          } on DioException catch (e) {
            return handler.next(e);
          }
        case RefreshResult.rejected:
          _client.onAuthExpired?.call();
        case RefreshResult.networkError:
          break;
      }
    }
    handler.next(err);
  }

  Future<RefreshResult> _tryRefresh() async {
    final refreshToken = await _storage.getRefreshToken();
    if (refreshToken == null) return RefreshResult.rejected;

    try {
      final response = await _refreshDioFactory().post(
        '${_dio.options.baseUrl}/api/v1/auth/refresh',
        data: {'refresh_token': refreshToken},
      );

      final accessToken = response.data['access_token'] as String;
      final newRefreshToken = response.data['refresh_token'] as String;
      await _storage.saveTokens(
        accessToken: accessToken,
        refreshToken: newRefreshToken,
      );
      return RefreshResult.refreshed;
    } on DioException catch (e) {
      if (e.response != null) {
        await _storage.clearTokens();
        return RefreshResult.rejected;
      }
      return RefreshResult.networkError;
    } catch (_) {
      await _storage.clearTokens();
      return RefreshResult.rejected;
    }
  }
}
