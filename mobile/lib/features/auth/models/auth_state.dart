import 'package:freezed_annotation/freezed_annotation.dart';

part 'auth_state.freezed.dart';

@freezed
sealed class AuthState with _$AuthState {
  const AuthState._();

  const factory AuthState.initial() = AuthInitial;
  const factory AuthState.loading() = AuthLoading;
  const factory AuthState.authenticated() = AuthAuthenticated;
  const factory AuthState.error(String message) = AuthError;

  bool get isAuthenticated => this is AuthAuthenticated;
  bool get isLoading => this is AuthLoading;
  String? get errorMessage => switch (this) {
        AuthError(message: final msg) => msg,
        _ => null,
      };
}
