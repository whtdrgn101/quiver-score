import 'package:flutter_test/flutter_test.dart';
import 'package:quiverscore/features/equipment/models/setup.dart';

void main() {
  group('SetupSummary', () {
    test('fromJson parses all fields', () {
      final json = {
        'id': 'setup-1',
        'name': 'Competition Recurve',
        'description': 'My main competition setup',
        'equipment_count': 5,
        'created_at': '2025-01-10T08:00:00Z',
      };

      final setup = SetupSummary.fromJson(json);
      expect(setup.id, 'setup-1');
      expect(setup.name, 'Competition Recurve');
      expect(setup.description, 'My main competition setup');
      expect(setup.equipmentCount, 5);
      expect(setup.createdAt, DateTime.utc(2025, 1, 10, 8));
    });

    test('fromJson handles null description', () {
      final json = {
        'id': 'setup-2',
        'name': 'Barebow',
        'description': null,
        'equipment_count': 0,
        'created_at': '2025-02-01T00:00:00Z',
      };

      final setup = SetupSummary.fromJson(json);
      expect(setup.description, isNull);
      expect(setup.equipmentCount, 0);
    });
  });

  group('SetupDetail', () {
    test('fromJson parses all fields including equipment', () {
      final json = {
        'id': 'setup-1',
        'name': 'Competition Recurve',
        'description': 'Full tournament setup',
        'brace_height': 8.75,
        'tiller': 0.125,
        'draw_weight': 38.0,
        'draw_length': 28.5,
        'arrow_foc': 12.5,
        'equipment': [
          {
            'id': 'eq-1',
            'category': 'riser',
            'name': 'Hoyt Formula Xi',
            'brand': 'Hoyt',
            'model': null,
            'specs': null,
            'notes': null,
            'created_at': '2025-01-01T00:00:00Z',
          }
        ],
        'created_at': '2025-01-10T08:00:00Z',
      };

      final detail = SetupDetail.fromJson(json);
      expect(detail.id, 'setup-1');
      expect(detail.braceHeight, 8.75);
      expect(detail.tiller, 0.125);
      expect(detail.drawWeight, 38.0);
      expect(detail.drawLength, 28.5);
      expect(detail.arrowFoc, 12.5);
      expect(detail.equipment.length, 1);
      expect(detail.equipment.first.name, 'Hoyt Formula Xi');
    });

    test('fromJson handles null tuning fields', () {
      final json = {
        'id': 'setup-2',
        'name': 'Barebow',
        'description': null,
        'brace_height': null,
        'tiller': null,
        'draw_weight': null,
        'draw_length': null,
        'arrow_foc': null,
        'equipment': [],
        'created_at': '2025-02-01T00:00:00Z',
      };

      final detail = SetupDetail.fromJson(json);
      expect(detail.braceHeight, isNull);
      expect(detail.tiller, isNull);
      expect(detail.drawWeight, isNull);
      expect(detail.drawLength, isNull);
      expect(detail.arrowFoc, isNull);
      expect(detail.equipment, isEmpty);
    });

    test('toJson includes only non-null tuning fields', () {
      final detail = SetupDetail(
        id: 'setup-1',
        name: 'Test Setup',
        description: 'Desc',
        drawWeight: 40.0,
        createdAt: DateTime(2025),
      );

      final json = detail.toJson();
      expect(json['name'], 'Test Setup');
      expect(json['description'], 'Desc');
      expect(json['draw_weight'], 40.0);
      expect(json.containsKey('brace_height'), false);
      expect(json.containsKey('tiller'), false);
    });
  });
}
