import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:quiverscore/core/database/database.dart';
import 'package:quiverscore/features/scoring/widgets/end_summary_row.dart';

EndsLocalData _makeEnd({int number = 1, int total = 27}) {
  return EndsLocalData(
    id: 'end-$number',
    sessionId: 'session-1',
    stageId: 'stage-1',
    endNumber: number,
    endTotal: total,
    serverId: null,
    createdAt: DateTime(2026, 1, 1),
  );
}

ArrowsLocalData _makeArrow(String endId, int num, String value, int score) {
  return ArrowsLocalData(
    id: 'arrow-$num',
    endId: endId,
    arrowNumber: num,
    scoreValue: value,
    scoreNumeric: score,
    xPos: null,
    yPos: null,
  );
}

Widget buildRow({
  EndsLocalData? end,
  List<ArrowsLocalData>? arrows,
  int imageCount = 0,
  bool imageSynced = true,
  VoidCallback? onImageTap,
  VoidCallback? onAddPhoto,
}) {
  final e = end ?? _makeEnd();
  final a = arrows ??
      [
        _makeArrow(e.id, 1, '9', 9),
        _makeArrow(e.id, 2, '9', 9),
        _makeArrow(e.id, 3, '9', 9),
      ];
  return MaterialApp(
    home: Scaffold(
      body: EndSummaryRow(
        end: e,
        arrows: a,
        imageCount: imageCount,
        imageSynced: imageSynced,
        onImageTap: onImageTap,
        onAddPhoto: onAddPhoto,
      ),
    ),
  );
}

void main() {
  testWidgets('displays end number and total', (tester) async {
    await tester.pumpWidget(buildRow(end: _makeEnd(number: 3, total: 27)));

    expect(find.text('3'), findsOneWidget);
    expect(find.text('27'), findsOneWidget);
  });

  testWidgets('displays arrow values', (tester) async {
    final end = _makeEnd();
    await tester.pumpWidget(buildRow(
      end: end,
      arrows: [
        _makeArrow(end.id, 1, 'X', 10),
        _makeArrow(end.id, 2, '10', 10),
        _makeArrow(end.id, 3, 'M', 0),
      ],
    ));

    expect(find.text('X'), findsOneWidget);
    expect(find.text('10'), findsOneWidget);
    expect(find.text('M'), findsOneWidget);
  });

  testWidgets('shows photo icon when imageCount > 0', (tester) async {
    await tester.pumpWidget(buildRow(imageCount: 1));

    expect(find.byIcon(Icons.photo), findsOneWidget);
  });

  testWidgets('shows cloud upload icon when image not synced', (tester) async {
    await tester.pumpWidget(buildRow(imageCount: 1, imageSynced: false));

    expect(find.byIcon(Icons.cloud_upload_outlined), findsOneWidget);
    expect(find.byIcon(Icons.photo), findsNothing);
  });

  testWidgets('no photo icon when imageCount is 0', (tester) async {
    await tester.pumpWidget(buildRow(imageCount: 0));

    expect(find.byIcon(Icons.photo), findsNothing);
    expect(find.byIcon(Icons.cloud_upload_outlined), findsNothing);
  });

  testWidgets('shows image count badge when multiple images', (tester) async {
    await tester.pumpWidget(buildRow(imageCount: 3));

    expect(find.text('3'), findsOneWidget);
  });

  testWidgets('tapping photo icon calls onImageTap', (tester) async {
    var tapped = false;
    await tester.pumpWidget(buildRow(
      imageCount: 1,
      onImageTap: () => tapped = true,
    ));

    await tester.tap(find.byIcon(Icons.photo));
    expect(tapped, isTrue);
  });

  testWidgets('shows add photo icon when onAddPhoto provided', (tester) async {
    var called = false;
    await tester.pumpWidget(buildRow(
      onAddPhoto: () => called = true,
    ));

    expect(find.byIcon(Icons.add_a_photo_outlined), findsOneWidget);
    await tester.tap(find.byIcon(Icons.add_a_photo_outlined));
    expect(called, isTrue);
  });

  testWidgets('displays various arrow colors without crashing', (tester) async {
    final end = _makeEnd(total: 45);
    await tester.pumpWidget(buildRow(
      end: end,
      arrows: [
        _makeArrow(end.id, 1, 'X', 10),
        _makeArrow(end.id, 2, '8', 8),
        _makeArrow(end.id, 3, '5', 5),
        _makeArrow(end.id, 4, '3', 3),
        _makeArrow(end.id, 5, '1', 1),
        _makeArrow(end.id, 6, 'M', 0),
      ],
    ));

    expect(find.text('45'), findsOneWidget);
  });
}
