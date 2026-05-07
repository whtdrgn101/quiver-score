import 'package:quiverscore/core/sync/sync_service.dart';

class FakeSyncService implements SyncService {
  int pullRoundTemplatesCallCount = 0;
  int syncPendingItemsCallCount = 0;

  @override
  Future<void> pullRoundTemplates() async {
    pullRoundTemplatesCallCount++;
  }

  @override
  Future<SyncResult> syncPendingItems() async {
    syncPendingItemsCallCount++;
    return SyncResult(synced: 0, failed: 0);
  }

  @override
  dynamic noSuchMethod(Invocation invocation) => super.noSuchMethod(invocation);
}
