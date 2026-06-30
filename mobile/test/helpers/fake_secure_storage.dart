import 'package:quiverscore/core/storage/secure_storage.dart';

class FakeSecureStorage implements SecureStorage {
  final Map<String, String> _data = {};

  @override
  Future<void> saveTokens({
    required String accessToken,
    required String refreshToken,
  }) async {
    _data['access_token'] = accessToken;
    _data['refresh_token'] = refreshToken;
  }

  @override
  Future<String?> getAccessToken() async => _data['access_token'];

  @override
  Future<String?> getRefreshToken() async => _data['refresh_token'];

  @override
  Future<void> saveUserId(String userId) async {
    _data['user_id'] = userId;
  }

  @override
  Future<String?> getUserId() async => _data['user_id'];

  @override
  Future<void> clearTokens() async => _data.clear();

  @override
  Future<void> setBiometricLockEnabled(bool enabled) async {
    _data['biometric_enabled'] = enabled ? 'true' : 'false';
  }

  @override
  Future<bool> isBiometricLockEnabled() async {
    return _data['biometric_enabled'] == 'true';
  }

  bool get isEmpty => _data.isEmpty;

  bool hasAccessToken() => _data.containsKey('access_token');
  bool hasRefreshToken() => _data.containsKey('refresh_token');
}
