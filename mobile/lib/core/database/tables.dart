import 'package:drift/drift.dart';

/// Round templates synced from the API (read-only reference data on mobile)
class RoundTemplates extends Table {
  TextColumn get id => text()();
  TextColumn get name => text()();
  TextColumn get organization => text()();
  TextColumn get description => text().nullable()();
  BoolColumn get isOfficial => boolean().withDefault(const Constant(false))();
  DateTimeColumn get syncedAt => dateTime()();

  @override
  Set<Column> get primaryKey => {id};
}

/// Stages belonging to a round template
class Stages extends Table {
  TextColumn get id => text()();
  TextColumn get templateId => text().references(RoundTemplates, #id)();
  TextColumn get name => text()();
  TextColumn get distance => text().nullable()();
  IntColumn get numEnds => integer()();
  IntColumn get arrowsPerEnd => integer()();
  TextColumn get allowedValues => text()(); // JSON array
  TextColumn get valueScoreMap => text()(); // JSON object
  IntColumn get maxScorePerArrow => integer()();
  IntColumn get stageOrder => integer()();

  @override
  Set<Column> get primaryKey => {id};
}

/// Scoring sessions created locally (offline-first)
class ScoringSessionsLocal extends Table {
  TextColumn get id => text()(); // client-generated UUID
  TextColumn get templateId => text().references(RoundTemplates, #id)();
  TextColumn get setupProfileId => text().nullable()();
  TextColumn get status =>
      text().withDefault(const Constant('in_progress'))();
  IntColumn get totalScore => integer().withDefault(const Constant(0))();
  IntColumn get totalXCount => integer().withDefault(const Constant(0))();
  IntColumn get totalArrows => integer().withDefault(const Constant(0))();
  TextColumn get notes => text().nullable()();
  TextColumn get location => text().nullable()();
  TextColumn get weather => text().nullable()();
  DateTimeColumn get startedAt => dateTime()();
  DateTimeColumn get completedAt => dateTime().nullable()();
  BoolColumn get synced => boolean().withDefault(const Constant(false))();
  TextColumn get serverId => text().nullable()(); // server-assigned ID after sync

  @override
  Set<Column> get primaryKey => {id};
}

/// Ends within a scoring session
class EndsLocal extends Table {
  TextColumn get id => text()();
  TextColumn get sessionId =>
      text().references(ScoringSessionsLocal, #id)();
  TextColumn get stageId => text()();
  IntColumn get endNumber => integer()();
  IntColumn get endTotal => integer().withDefault(const Constant(0))();
  DateTimeColumn get createdAt => dateTime()();

  @override
  Set<Column> get primaryKey => {id};
}

/// Individual arrows within an end
class ArrowsLocal extends Table {
  TextColumn get id => text()();
  TextColumn get endId => text().references(EndsLocal, #id)();
  IntColumn get arrowNumber => integer()();
  TextColumn get scoreValue => text()();
  IntColumn get scoreNumeric => integer()();
  RealColumn get xPos => real().nullable()();
  RealColumn get yPos => real().nullable()();

  @override
  Set<Column> get primaryKey => {id};
}

/// Photos taken of targets per end
class EndImages extends Table {
  TextColumn get id => text()();
  TextColumn get endId => text().references(EndsLocal, #id)();
  TextColumn get filePath => text()(); // local file path
  DateTimeColumn get capturedAt => dateTime()();
  BoolColumn get synced => boolean().withDefault(const Constant(false))();

  @override
  Set<Column> get primaryKey => {id};
}

/// Queue of mutations to sync with the API when online
class SyncQueue extends Table {
  IntColumn get id => integer().autoIncrement()();
  TextColumn get entityType => text()(); // session, end, image
  TextColumn get entityId => text()();
  TextColumn get action => text()(); // create, update, complete, delete
  TextColumn get payloadJson => text()();
  DateTimeColumn get createdAt => dateTime()();
  DateTimeColumn get syncedAt => dateTime().nullable()();
  IntColumn get retryCount => integer().withDefault(const Constant(0))();
  TextColumn get lastError => text().nullable()();
}
