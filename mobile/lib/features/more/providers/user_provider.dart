import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/api_client.dart';

final currentUserProvider = FutureProvider<UserInfo>((ref) async {
  final api = ref.read(apiClientProvider);
  final response = await api.dio.get('/api/v1/users/me');
  return UserInfo.fromJson(response.data as Map<String, dynamic>);
});

class UserInfo {
  final String id;
  final String email;
  final String username;
  final String? displayName;
  final String? bowType;

  const UserInfo({
    required this.id,
    required this.email,
    required this.username,
    this.displayName,
    this.bowType,
  });

  factory UserInfo.fromJson(Map<String, dynamic> json) {
    return UserInfo(
      id: json['id'] as String,
      email: json['email'] as String,
      username: json['username'] as String,
      displayName: json['display_name'] as String?,
      bowType: json['bow_type'] as String?,
    );
  }
}
