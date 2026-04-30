import 'package:flutter_test/flutter_test.dart';
import 'package:quiverscore/features/scoring/providers/scoring_provider.dart';

void main() {
  group('ScoringState', () {
    test('default state has no active session', () {
      const state = ScoringState();
      expect(state.activeSession, isNull);
      expect(state.ends, isEmpty);
      expect(state.arrowsByEnd, isEmpty);
      expect(state.currentStage, isNull);
    });

    test('copyWith preserves existing values when null passed', () {
      const state = ScoringState();
      final copy = state.copyWith();
      expect(copy.activeSession, isNull);
      expect(copy.ends, isEmpty);
      expect(copy.arrowsByEnd, isEmpty);
    });

    test('copyWith replaces provided values', () {
      const state = ScoringState();
      final copy = state.copyWith(ends: []);
      expect(copy.ends, isEmpty);
    });
  });

  group('ArrowInput', () {
    test('stores score value', () {
      const arrow = ArrowInput(scoreValue: 'X');
      expect(arrow.scoreValue, 'X');
      expect(arrow.xPos, isNull);
      expect(arrow.yPos, isNull);
    });

    test('stores position data', () {
      const arrow = ArrowInput(scoreValue: '10', xPos: 0.5, yPos: -0.3);
      expect(arrow.scoreValue, '10');
      expect(arrow.xPos, 0.5);
      expect(arrow.yPos, -0.3);
    });

    test('miss value', () {
      const arrow = ArrowInput(scoreValue: 'M');
      expect(arrow.scoreValue, 'M');
    });
  });
}
