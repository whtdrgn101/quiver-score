import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/database/database.dart';

/// Stream of all cached round templates from local DB
final roundTemplatesProvider =
    StreamProvider<List<RoundTemplate>>((ref) {
  final db = ref.watch(databaseProvider);
  return (db.select(db.roundTemplates)
        ..orderBy([(t) => OrderingTerm.asc(t.name)]))
      .watch();
});

/// Get stages for a specific template
final stagesForTemplateProvider =
    FutureProvider.family<List<StagesCompanion>, String>((ref, templateId) async {
  final db = ref.watch(databaseProvider);
  final rows = await (db.select(db.stages)
        ..where((t) => t.templateId.equals(templateId))
        ..orderBy([(t) => OrderingTerm.asc(t.stageOrder)]))
      .get();
  return rows
      .map((s) => StagesCompanion.insert(
            id: s.id,
            templateId: s.templateId,
            name: s.name,
            distance: Value(s.distance),
            numEnds: s.numEnds,
            arrowsPerEnd: s.arrowsPerEnd,
            allowedValues: s.allowedValues,
            valueScoreMap: s.valueScoreMap,
            maxScorePerArrow: s.maxScorePerArrow,
            stageOrder: s.stageOrder,
          ))
      .toList();
});
