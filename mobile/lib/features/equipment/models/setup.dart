import 'equipment.dart';

class SetupSummary {
  final String id;
  final String name;
  final String? description;
  final int equipmentCount;
  final DateTime createdAt;

  const SetupSummary({
    required this.id,
    required this.name,
    this.description,
    required this.equipmentCount,
    required this.createdAt,
  });

  factory SetupSummary.fromJson(Map<String, dynamic> json) {
    return SetupSummary(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      equipmentCount: json['equipment_count'] as int? ?? 0,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }
}

class SetupDetail {
  final String id;
  final String name;
  final String? description;
  final double? braceHeight;
  final double? tiller;
  final double? drawWeight;
  final double? drawLength;
  final double? arrowFoc;
  final List<Equipment> equipment;
  final DateTime createdAt;

  const SetupDetail({
    required this.id,
    required this.name,
    this.description,
    this.braceHeight,
    this.tiller,
    this.drawWeight,
    this.drawLength,
    this.arrowFoc,
    this.equipment = const [],
    required this.createdAt,
  });

  factory SetupDetail.fromJson(Map<String, dynamic> json) {
    return SetupDetail(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      braceHeight: (json['brace_height'] as num?)?.toDouble(),
      tiller: (json['tiller'] as num?)?.toDouble(),
      drawWeight: (json['draw_weight'] as num?)?.toDouble(),
      drawLength: (json['draw_length'] as num?)?.toDouble(),
      arrowFoc: (json['arrow_foc'] as num?)?.toDouble(),
      equipment: (json['equipment'] as List?)
              ?.map((e) => Equipment.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  Map<String, dynamic> toJson() => {
        'name': name,
        if (description != null) 'description': description,
        if (braceHeight != null) 'brace_height': braceHeight,
        if (tiller != null) 'tiller': tiller,
        if (drawWeight != null) 'draw_weight': drawWeight,
        if (drawLength != null) 'draw_length': drawLength,
        if (arrowFoc != null) 'arrow_foc': arrowFoc,
      };
}
