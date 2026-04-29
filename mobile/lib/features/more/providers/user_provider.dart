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
  final String? classification;
  final String? bio;
  final String? avatar;
  final bool profilePublic;
  final Map<String, dynamic>? socialLinks;

  const UserInfo({
    required this.id,
    required this.email,
    required this.username,
    this.displayName,
    this.bowType,
    this.classification,
    this.bio,
    this.avatar,
    this.profilePublic = false,
    this.socialLinks,
  });

  factory UserInfo.fromJson(Map<String, dynamic> json) {
    return UserInfo(
      id: json['id'] as String,
      email: json['email'] as String,
      username: json['username'] as String,
      displayName: json['display_name'] as String?,
      bowType: json['bow_type'] as String?,
      classification: json['classification'] as String?,
      bio: json['bio'] as String?,
      avatar: json['avatar'] as String?,
      profilePublic: json['profile_public'] as bool? ?? false,
      socialLinks: json['social_links'] as Map<String, dynamic>?,
    );
  }
}
