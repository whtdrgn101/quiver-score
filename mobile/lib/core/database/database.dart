import 'package:drift/drift.dart';
import 'package:drift/native.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path_provider/path_provider.dart';
import 'package:path/path.dart' as p;
import 'dart:io';

import 'tables.dart';

part 'database.g.dart';

final databaseProvider = Provider<AppDatabase>((ref) {
  throw UnimplementedError('Database must be overridden in main.dart');
});

@DriftDatabase(tables: [
  RoundTemplates,
  Stages,
  ScoringSessionsLocal,
  EndsLocal,
  ArrowsLocal,
  EndImages,
  SyncQueue,
  EquipmentCache,
  SetupCache,
  SetupEquipmentCache,
  ClubCache,
])
class AppDatabase extends _$AppDatabase {
  AppDatabase() : super(_openConnection());

  @override
  int get schemaVersion => 3;

  @override
  MigrationStrategy get migration => MigrationStrategy(
        onUpgrade: (migrator, from, to) async {
          if (from < 2) {
            await customStatement(
                'ALTER TABLE ends_local ADD COLUMN server_id TEXT');
          }
          if (from < 3) {
            await migrator.createTable(equipmentCache);
            await migrator.createTable(setupCache);
            await migrator.createTable(setupEquipmentCache);
            await migrator.createTable(clubCache);
          }
        },
      );

  static LazyDatabase _openConnection() {
    return LazyDatabase(() async {
      final dbFolder = await getApplicationDocumentsDirectory();
      final file = File(p.join(dbFolder.path, 'quiverscore.db'));
      return NativeDatabase.createInBackground(file);
    });
  }
}
