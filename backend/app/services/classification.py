"""Classification service with lookup tables for ArcheryGB and NFAA systems."""

# ArcheryGB classifications for Recurve (simplified thresholds)
# Format: {round_name: [(score_threshold, classification), ...]} ordered highest first
ARCHERY_GB_RECURVE = {
    "WA 720 (70m)": [
        (625, "Grand Master Bowman"),
        (575, "Master Bowman"),
        (525, "Bowman 1st Class"),
        (475, "Bowman 2nd Class"),
        (400, "Bowman 3rd Class"),
        (300, "Archer 1st Class"),
        (200, "Archer 2nd Class"),
        (100, "Archer 3rd Class"),
    ],
    "WA 720 (60m)": [
        (640, "Grand Master Bowman"),
        (590, "Master Bowman"),
        (540, "Bowman 1st Class"),
        (490, "Bowman 2nd Class"),
        (420, "Bowman 3rd Class"),
        (320, "Archer 1st Class"),
        (220, "Archer 2nd Class"),
        (120, "Archer 3rd Class"),
    ],
    "WA 18m Round (60 arrows)": [
        (550, "Grand Master Bowman"),
        (510, "Master Bowman"),
        (470, "Bowman 1st Class"),
        (420, "Bowman 2nd Class"),
        (350, "Bowman 3rd Class"),
        (270, "Archer 1st Class"),
        (180, "Archer 2nd Class"),
        (90, "Archer 3rd Class"),
    ],
}

# NFAA classifications for indoor/outdoor rounds
NFAA_CLASSIFICATIONS = {
    "NFAA 300 Indoor": [
        (290, "Expert"),
        (270, "Sharpshooter"),
        (240, "Marksman"),
        (200, "Bowman"),
    ],
    "NFAA 300 Outdoor": [
        (280, "Expert"),
        (260, "Sharpshooter"),
        (230, "Marksman"),
        (190, "Bowman"),
    ],
}

# Combined lookup
ALL_CLASSIFICATIONS = {
    **{k: ("ArcheryGB", v) for k, v in ARCHERY_GB_RECURVE.items()},
    **{k: ("NFAA", v) for k, v in NFAA_CLASSIFICATIONS.items()},
}


def calculate_classification(score: int, template_name: str) -> tuple[str, str] | None:
    """Calculate classification for a given score and round type.

    Returns (system, classification) or None if no classification applies.
    """
    entry = ALL_CLASSIFICATIONS.get(template_name)
    if not entry:
        return None

    system, thresholds = entry
    for threshold, classification in thresholds:
        if score >= threshold:
            return (system, classification)

    return None
