import 'dart:convert';

import 'package:flutter/material.dart';

class Equipment {
  final String id;
  final String category;
  final String name;
  final String? brand;
  final String? model;
  final Map<String, dynamic>? specs;
  final String? notes;
  final DateTime createdAt;

  const Equipment({
    required this.id,
    required this.category,
    required this.name,
    this.brand,
    this.model,
    this.specs,
    this.notes,
    required this.createdAt,
  });

  factory Equipment.fromJson(Map<String, dynamic> json) {
    return Equipment(
      id: json['id'] as String,
      category: json['category'] as String,
      name: json['name'] as String,
      brand: json['brand'] as String?,
      model: json['model'] as String?,
      specs: json['specs'] is Map
          ? Map<String, dynamic>.from(json['specs'] as Map)
          : json['specs'] is String
              ? jsonDecode(json['specs'] as String) as Map<String, dynamic>?
              : null,
      notes: json['notes'] as String?,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  Map<String, dynamic> toJson() => {
        'category': category,
        'name': name,
        if (brand != null) 'brand': brand,
        if (model != null) 'model': model,
        if (specs != null) 'specs': specs,
        if (notes != null) 'notes': notes,
      };

  static const categories = [
    'riser',
    'limbs',
    'arrows',
    'sight',
    'stabilizer',
    'rest',
    'release',
    'scope',
    'string',
    'other',
  ];

  static const categoryLabels = {
    'riser': 'Riser',
    'limbs': 'Limbs',
    'arrows': 'Arrows',
    'sight': 'Sight',
    'stabilizer': 'Stabilizer',
    'rest': 'Rest',
    'release': 'Release',
    'scope': 'Scope',
    'string': 'String',
    'other': 'Other',
  };

  static const categoryIcons = {
    'riser': Icons.sports_outlined,
    'limbs': Icons.linear_scale,
    'arrows': Icons.arrow_forward,
    'sight': Icons.gps_fixed,
    'stabilizer': Icons.balance,
    'rest': Icons.airline_seat_recline_normal,
    'release': Icons.touch_app,
    'scope': Icons.circle_outlined,
    'string': Icons.cable,
    'other': Icons.build_outlined,
  };

  String get categoryLabel => categoryLabels[category] ?? category;
  IconData get categoryIcon => categoryIcons[category] ?? Icons.build_outlined;
}

class EquipmentStats {
  final String itemId;
  final String itemName;
  final String category;
  final int sessionsCount;
  final int totalArrows;
  final DateTime? lastUsed;

  const EquipmentStats({
    required this.itemId,
    required this.itemName,
    required this.category,
    required this.sessionsCount,
    required this.totalArrows,
    this.lastUsed,
  });

  factory EquipmentStats.fromJson(Map<String, dynamic> json) {
    return EquipmentStats(
      itemId: json['item_id'] as String,
      itemName: json['item_name'] as String,
      category: json['category'] as String,
      sessionsCount: json['sessions_count'] as int? ?? 0,
      totalArrows: json['total_arrows'] as int? ?? 0,
      lastUsed: json['last_used'] != null
          ? DateTime.parse(json['last_used'] as String)
          : null,
    );
  }
}
