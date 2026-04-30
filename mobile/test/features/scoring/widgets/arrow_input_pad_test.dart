import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:quiverscore/features/scoring/widgets/arrow_input_pad.dart';

void main() {
  Widget buildPad({
    List<String> allowedValues = const ['X', '10', '9', '8', '7', '6', '5', '4', '3', '2', '1', 'M'],
    ValueChanged<String>? onValueTap,
    VoidCallback? onBackspace,
    VoidCallback? onSubmit,
  }) {
    return MaterialApp(
      home: Scaffold(
        body: ArrowInputPad(
          allowedValues: allowedValues,
          onValueTap: onValueTap ?? (_) {},
          onBackspace: onBackspace ?? () {},
          onSubmit: onSubmit,
        ),
      ),
    );
  }

  testWidgets('renders all allowed values as buttons', (tester) async {
    final values = ['X', '10', '9', 'M'];
    await tester.pumpWidget(buildPad(allowedValues: values));

    for (final v in values) {
      expect(find.text(v), findsOneWidget);
    }
  });

  testWidgets('tapping a value calls onValueTap with correct value', (tester) async {
    String? tapped;
    await tester.pumpWidget(buildPad(
      allowedValues: ['X', '10', '9'],
      onValueTap: (v) => tapped = v,
    ));

    await tester.tap(find.text('X'));
    expect(tapped, 'X');

    await tester.tap(find.text('10'));
    expect(tapped, '10');
  });

  testWidgets('tapping Undo calls onBackspace', (tester) async {
    var called = false;
    await tester.pumpWidget(buildPad(onBackspace: () => called = true));

    await tester.tap(find.text('Undo'));
    expect(called, isTrue);
  });

  testWidgets('Submit End button is disabled when onSubmit is null', (tester) async {
    await tester.pumpWidget(buildPad(onSubmit: null));

    final button = tester.widget<FilledButton>(find.widgetWithText(FilledButton, 'Submit End'));
    expect(button.onPressed, isNull);
  });

  testWidgets('Submit End button is enabled and callable when onSubmit provided', (tester) async {
    var submitted = false;
    await tester.pumpWidget(buildPad(onSubmit: () => submitted = true));

    await tester.tap(find.text('Submit End'));
    expect(submitted, isTrue);
  });

  testWidgets('renders with minimal values (e.g. 6-zone)', (tester) async {
    final values = ['6', '5', '4', '3', '2', '1'];
    await tester.pumpWidget(buildPad(allowedValues: values));

    for (final v in values) {
      expect(find.text(v), findsOneWidget);
    }
    expect(find.text('X'), findsNothing);
  });
}
