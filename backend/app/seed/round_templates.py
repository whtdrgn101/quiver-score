from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.models.round_template import RoundTemplate, RoundTemplateStage

TEMPLATES = [
    {
        "name": "WA Indoor 18m (Recurve)",
        "organization": "WA",
        "description": "World Archery Indoor round: 60 arrows at 18m on 40cm target face, 10-ring scoring with X",
        "stages": [
            {
                "stage_order": 1,
                "name": "18m",
                "distance": "18m",
                "num_ends": 20,
                "arrows_per_end": 3,
                "allowed_values": ["X", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "M"],
                "value_score_map": {"X": 10, "10": 10, "9": 9, "8": 8, "7": 7, "6": 6, "5": 5, "4": 4, "3": 3, "2": 2, "1": 1, "M": 0},
                "max_score_per_arrow": 10,
            }
        ],
    },
    {
        "name": "WA Indoor 18m (Compound)",
        "organization": "WA",
        "description": "World Archery Indoor Compound: 60 arrows at 18m on triple spot, inner-10 scoring with X",
        "stages": [
            {
                "stage_order": 1,
                "name": "18m Triple",
                "distance": "18m",
                "num_ends": 20,
                "arrows_per_end": 3,
                "allowed_values": ["X", "10", "9", "8", "7", "6", "M"],
                "value_score_map": {"X": 10, "10": 10, "9": 9, "8": 8, "7": 7, "6": 6, "M": 0},
                "max_score_per_arrow": 10,
            }
        ],
    },
    {
        "name": "WA 720 (70m Recurve)",
        "organization": "WA",
        "description": "World Archery 720 round: 72 arrows at 70m on 122cm target face",
        "stages": [
            {
                "stage_order": 1,
                "name": "70m",
                "distance": "70m",
                "num_ends": 12,
                "arrows_per_end": 6,
                "allowed_values": ["X", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "M"],
                "value_score_map": {"X": 10, "10": 10, "9": 9, "8": 8, "7": 7, "6": 6, "5": 5, "4": 4, "3": 3, "2": 2, "1": 1, "M": 0},
                "max_score_per_arrow": 10,
            }
        ],
    },
    {
        "name": "Vegas 300",
        "organization": "Vegas",
        "description": "The Vegas Shoot: 30 arrows at 20yd on 40cm target, inner X ring",
        "stages": [
            {
                "stage_order": 1,
                "name": "20yd",
                "distance": "20yd",
                "num_ends": 10,
                "arrows_per_end": 3,
                "allowed_values": ["X", "10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "M"],
                "value_score_map": {"X": 10, "10": 10, "9": 9, "8": 8, "7": 7, "6": 6, "5": 5, "4": 4, "3": 3, "2": 2, "1": 1, "M": 0},
                "max_score_per_arrow": 10,
            }
        ],
    },
    {
        "name": "NFAA Indoor 300",
        "organization": "NFAA",
        "description": "NFAA Indoor round: 60 arrows at 20yd, 5-ring blue target face, X=5",
        "stages": [
            {
                "stage_order": 1,
                "name": "20yd",
                "distance": "20yd",
                "num_ends": 12,
                "arrows_per_end": 5,
                "allowed_values": ["X", "5", "4", "3", "2", "1", "M"],
                "value_score_map": {"X": 5, "5": 5, "4": 4, "3": 3, "2": 2, "1": 1, "M": 0},
                "max_score_per_arrow": 5,
            }
        ],
    },
]


async def seed_round_templates(db: AsyncSession):
    result = await db.execute(select(RoundTemplate).limit(1))
    if result.scalar_one_or_none():
        return  # already seeded

    for tpl in TEMPLATES:
        template = RoundTemplate(
            name=tpl["name"],
            organization=tpl["organization"],
            description=tpl["description"],
            is_official=True,
        )
        db.add(template)
        await db.flush()

        for stage_data in tpl["stages"]:
            stage = RoundTemplateStage(template_id=template.id, **stage_data)
            db.add(stage)

    await db.commit()
