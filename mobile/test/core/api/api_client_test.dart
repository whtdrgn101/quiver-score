import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:quiverscore/core/api/api_client.dart';

import '../../helpers/fake_secure_storage.dart';

class _MockHttpAdapter implements HttpClientAdapter {
  final Future<ResponseBody> Function(RequestOptions options) handler;

  _MockHttpAdapter(this.handler);

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) => handler(options);

  @override
  void close({bool force = false}) {}
}

Dio _mockDioWithAdapter(
    Future<ResponseBody> Function(RequestOptions) handler) {
  final dio = Dio();
  dio.httpClientAdapter = _MockHttpAdapter(handler);
  return dio;
}

ResponseBody _jsonResponse(Map<String, dynamic> data, {int statusCode = 200}) {
  final json = data.entries.map((e) => '"${e.key}":"${e.value}"').join(',');
  return ResponseBody.fromString(
    '{$json}',
    statusCode,
    headers: {
      'content-type': ['application/json'],
    },
  );
}

void main() {
  group('RefreshResult enum', () {
    test('has three values', () {
      expect(RefreshResult.values.length, 3);
      expect(RefreshResult.values,
          contains(RefreshResult.refreshed));
      expect(RefreshResult.values,
          contains(RefreshResult.rejected));
      expect(RefreshResult.values,
          contains(RefreshResult.networkError));
    });
  });

  group('ApiClient interceptor', () {
    late FakeSecureStorage storage;
    late ApiClient client;
    late int authExpiredCallCount;

    setUp(() {
      storage = FakeSecureStorage();
      authExpiredCallCount = 0;
    });

    test('attaches access token to requests', () async {
      await storage.saveTokens(
        accessToken: 'test-access-token',
        refreshToken: 'test-refresh-token',
      );

      client = ApiClient(storage: storage, baseUrl: 'https://test.com');
      client.dio.httpClientAdapter = _MockHttpAdapter((options) async {
        expect(options.headers['Authorization'],
            'Bearer test-access-token');
        return _jsonResponse({'ok': 'true'});
      });

      final response = await client.dio.get('/api/v1/test');
      expect(response.statusCode, 200);
    });

    test('sends requests without auth header when no token stored', () async {
      client = ApiClient(storage: storage, baseUrl: 'https://test.com');
      client.dio.httpClientAdapter = _MockHttpAdapter((options) async {
        expect(options.headers['Authorization'], isNull);
        return _jsonResponse({'ok': 'true'});
      });

      final response = await client.dio.get('/api/v1/test');
      expect(response.statusCode, 200);
    });

    test('refreshes token on 401 and retries request', () async {
      await storage.saveTokens(
        accessToken: 'expired-token',
        refreshToken: 'valid-refresh-token',
      );

      var requestCount = 0;

      final refreshDio = _mockDioWithAdapter((options) async {
        expect(options.uri.path, '/api/v1/auth/refresh');
        return _jsonResponse({
          'access_token': 'new-access-token',
          'refresh_token': 'new-refresh-token',
        });
      });

      client = ApiClient(
        storage: storage,
        baseUrl: 'https://test.com',
        refreshDioFactory: () => refreshDio,
      );
      client.onAuthExpired = () => authExpiredCallCount++;

      client.dio.httpClientAdapter = _MockHttpAdapter((options) async {
        requestCount++;
        if (requestCount == 1) {
          return _jsonResponse({'detail': 'Not authenticated'},
              statusCode: 401);
        }
        // Retry should have new token
        expect(options.headers['Authorization'],
            'Bearer new-access-token');
        return _jsonResponse({'ok': 'true'});
      });

      final response = await client.dio.get('/api/v1/test');
      expect(response.statusCode, 200);
      expect(requestCount, 2);
      expect(authExpiredCallCount, 0);
      expect(await storage.getAccessToken(), 'new-access-token');
      expect(await storage.getRefreshToken(), 'new-refresh-token');
    });

    test('calls onAuthExpired when server rejects refresh token', () async {
      await storage.saveTokens(
        accessToken: 'expired-token',
        refreshToken: 'expired-refresh-token',
      );

      final refreshDio = _mockDioWithAdapter((options) async {
        return _jsonResponse({'detail': 'Token expired'}, statusCode: 401);
      });

      client = ApiClient(
        storage: storage,
        baseUrl: 'https://test.com',
        refreshDioFactory: () => refreshDio,
      );
      client.onAuthExpired = () => authExpiredCallCount++;

      client.dio.httpClientAdapter = _MockHttpAdapter((options) async {
        return _jsonResponse({'detail': 'Not authenticated'},
            statusCode: 401);
      });

      try {
        await client.dio.get('/api/v1/test');
      } on DioException catch (e) {
        expect(e.response?.statusCode, 401);
      }

      expect(authExpiredCallCount, 1);
      expect(storage.isEmpty, true);
    });

    test('does NOT call onAuthExpired on network error during refresh',
        () async {
      await storage.saveTokens(
        accessToken: 'expired-token',
        refreshToken: 'valid-refresh-token',
      );

      final refreshDio = _mockDioWithAdapter((options) async {
        throw DioException(
          type: DioExceptionType.connectionError,
          requestOptions: options,
          message: 'No internet',
        );
      });

      client = ApiClient(
        storage: storage,
        baseUrl: 'https://test.com',
        refreshDioFactory: () => refreshDio,
      );
      client.onAuthExpired = () => authExpiredCallCount++;

      client.dio.httpClientAdapter = _MockHttpAdapter((options) async {
        return _jsonResponse({'detail': 'Not authenticated'},
            statusCode: 401);
      });

      try {
        await client.dio.get('/api/v1/test');
      } on DioException catch (e) {
        expect(e.response?.statusCode, 401);
      }

      expect(authExpiredCallCount, 0);
      expect(storage.hasAccessToken(), true);
      expect(storage.hasRefreshToken(), true);
    });

    test('does NOT call onAuthExpired on connection timeout during refresh',
        () async {
      await storage.saveTokens(
        accessToken: 'expired-token',
        refreshToken: 'valid-refresh-token',
      );

      final refreshDio = _mockDioWithAdapter((options) async {
        throw DioException(
          type: DioExceptionType.connectionTimeout,
          requestOptions: options,
          message: 'Timed out',
        );
      });

      client = ApiClient(
        storage: storage,
        baseUrl: 'https://test.com',
        refreshDioFactory: () => refreshDio,
      );
      client.onAuthExpired = () => authExpiredCallCount++;

      client.dio.httpClientAdapter = _MockHttpAdapter((options) async {
        return _jsonResponse({'detail': 'Not authenticated'},
            statusCode: 401);
      });

      try {
        await client.dio.get('/api/v1/test');
      } on DioException catch (e) {
        expect(e.response?.statusCode, 401);
      }

      expect(authExpiredCallCount, 0);
      expect(storage.hasRefreshToken(), true);
    });

    test('calls onAuthExpired when no refresh token stored', () async {
      await storage.saveTokens(
        accessToken: 'expired-token',
        refreshToken: 'test',
      );
      // Clear to simulate missing refresh token
      await storage.clearTokens();
      // Re-add just the access token scenario isn't possible with
      // FakeSecureStorage, so test with fully empty storage

      client = ApiClient(
        storage: storage,
        baseUrl: 'https://test.com',
      );
      client.onAuthExpired = () => authExpiredCallCount++;

      client.dio.httpClientAdapter = _MockHttpAdapter((options) async {
        return _jsonResponse({'detail': 'Not authenticated'},
            statusCode: 401);
      });

      try {
        await client.dio.get('/api/v1/test');
      } on DioException catch (e) {
        expect(e.response?.statusCode, 401);
      }

      expect(authExpiredCallCount, 1);
    });

    test('non-401 errors pass through without refresh attempt', () async {
      await storage.saveTokens(
        accessToken: 'valid-token',
        refreshToken: 'valid-refresh',
      );

      var refreshAttempted = false;

      final refreshDio = _mockDioWithAdapter((options) async {
        refreshAttempted = true;
        return _jsonResponse({'error': 'should not be called'});
      });

      client = ApiClient(
        storage: storage,
        baseUrl: 'https://test.com',
        refreshDioFactory: () => refreshDio,
      );

      client.dio.httpClientAdapter = _MockHttpAdapter((options) async {
        return _jsonResponse({'detail': 'Not found'}, statusCode: 404);
      });

      try {
        await client.dio.get('/api/v1/test');
      } on DioException catch (e) {
        expect(e.response?.statusCode, 404);
      }

      expect(refreshAttempted, false);
      expect(storage.hasAccessToken(), true);
    });
  });
}
