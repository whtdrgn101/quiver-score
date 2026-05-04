import 'package:flutter_test/flutter_test.dart';
import 'package:quiverscore/features/equipment/models/equipment.dart';

void main() {
  group('Equipment', () {
    test('fromJson parses all fields', () {
      final json = {
        'id': 'eq-1',
        'category': 'riser',
        'name': 'Hoyt Formula Xi',
        'brand': 'Hoyt',
        'model': 'Formula Xi',
        'specs': {'color': 'red'},
        'notes': 'Primary riser',
        'created_at': '2025-01-15T10:30:00Z',
      };

      final eq = Equipment.fromJson(json);
      expect(eq.id, 'eq-1');
      expect(eq.category, 'riser');
      expect(eq.name, 'Hoyt Formula Xi');
      expect(eq.brand, 'Hoyt');
      expect(eq.model, 'Formula Xi');
      expect(eq.specs, {'color': 'red'});
      expect(eq.notes, 'Primary riser');
      expect(eq.createdAt, DateTime.utc(2025, 1, 15, 10, 30));
    });

    test('fromJson handles null optional fields', () {
      final json = {
        'id': 'eq-2',
        'category': 'arrows',
        'name': 'X10',
        'brand': null,
        'model': null,
        'specs': null,
        'notes': null,
        'created_at': '2025-02-01T00:00:00Z',
      };

      final eq = Equipment.fromJson(json);
      expect(eq.brand, isNull);
      expect(eq.model, isNull);
      expect(eq.specs, isNull);
      expect(eq.notes, isNull);
    });

    test('toJson includes only non-null fields', () {
      final json = Equipment(
        id: 'eq-1',
        category: 'sight',
        name: 'Shibuya Ultima',
        brand: 'Shibuya',
        createdAt: DateTime(2025),
      ).toJson();
      expect(json['category'], 'sight');
      expect(json['name'], 'Shibuya Ultima');
      expect(json['brand'], 'Shibuya');
      expect(json.containsKey('model'), false);
      expect(json.containsKey('notes'), false);
    });

    test('categoryLabel returns correct display name', () {
      final item = Equipment(
        id: 'eq-1',
        category: 'stabilizer',
        name: 'Test',
        createdAt: DateTime(2025),
      );
      expect(item.categoryLabel, 'Stabilizer');
    });

    test('categories list has 10 items', () {
      expect(Equipment.categories.length, 10);
      expect(Equipment.categories, contains('riser'));
      expect(Equipment.categories, contains('other'));
    });
  });

  group('EquipmentStats', () {
    test('fromJson parses all fields', () {
      final json = {
        'item_id': 'eq-1',
        'item_name': 'Hoyt Formula Xi',
        'category': 'riser',
        'sessions_count': 42,
        'total_arrows': 1260,
        'last_used': '2025-03-01T14:00:00Z',
      };

      final stats = EquipmentStats.fromJson(json);
      expect(stats.itemId, 'eq-1');
      expect(stats.itemName, 'Hoyt Formula Xi');
      expect(stats.category, 'riser');
      expect(stats.sessionsCount, 42);
      expect(stats.totalArrows, 1260);
      expect(stats.lastUsed, DateTime.utc(2025, 3, 1, 14));
    });

    test('fromJson handles null last_used', () {
      final json = {
        'item_id': 'eq-2',
        'item_name': 'New Arrows',
        'category': 'arrows',
        'sessions_count': 0,
        'total_arrows': 0,
        'last_used': null,
      };

      final stats = EquipmentStats.fromJson(json);
      expect(stats.lastUsed, isNull);
      expect(stats.sessionsCount, 0);
    });
  });
}
