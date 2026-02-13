import csv
import io
from datetime import datetime

from reportlab.lib import colors
from reportlab.lib.pagesizes import letter
from reportlab.lib.units import inch
from reportlab.platypus import SimpleDocTemplate, Table, TableStyle, Paragraph, Spacer
from reportlab.lib.styles import getSampleStyleSheet

from app.models.scoring import ScoringSession


def generate_session_csv(session: ScoringSession) -> str:
    output = io.StringIO()
    writer = csv.writer(output)

    template_name = session.template.name if session.template else "Unknown"
    writer.writerow(["Session Detail"])
    writer.writerow(["Round", template_name])
    writer.writerow(["Status", session.status])
    writer.writerow(["Total Score", session.total_score])
    writer.writerow(["Total X Count", session.total_x_count])
    writer.writerow(["Total Arrows", session.total_arrows])
    writer.writerow(["Location", session.location or ""])
    writer.writerow(["Weather", session.weather or ""])
    writer.writerow(["Notes", session.notes or ""])
    writer.writerow(["Started", session.started_at.isoformat() if session.started_at else ""])
    writer.writerow(["Completed", session.completed_at.isoformat() if session.completed_at else ""])
    writer.writerow([])

    writer.writerow(["End", "Arrow", "Value", "Score"])
    for end in session.ends:
        for arrow in end.arrows:
            writer.writerow([end.end_number, arrow.arrow_number, arrow.score_value, arrow.score_numeric])

    return output.getvalue()


def generate_sessions_csv(sessions: list[ScoringSession]) -> str:
    output = io.StringIO()
    writer = csv.writer(output)

    writer.writerow(["Date", "Round", "Status", "Score", "X Count", "Arrows", "Location", "Notes"])
    for s in sessions:
        template_name = s.template.name if s.template else "Unknown"
        date_str = (s.completed_at or s.started_at).strftime("%Y-%m-%d") if (s.completed_at or s.started_at) else ""
        writer.writerow([
            date_str,
            template_name,
            s.status,
            s.total_score,
            s.total_x_count,
            s.total_arrows,
            s.location or "",
            s.notes or "",
        ])

    return output.getvalue()


def generate_session_pdf(session: ScoringSession) -> bytes:
    buffer = io.BytesIO()
    doc = SimpleDocTemplate(buffer, pagesize=letter, topMargin=0.5 * inch)
    styles = getSampleStyleSheet()
    elements = []

    template_name = session.template.name if session.template else "Unknown"
    elements.append(Paragraph(f"QuiverScore — {template_name}", styles["Title"]))
    elements.append(Spacer(1, 12))

    info_data = [
        ["Status", session.status],
        ["Total Score", str(session.total_score)],
        ["X Count", str(session.total_x_count)],
        ["Total Arrows", str(session.total_arrows)],
    ]
    if session.location:
        info_data.append(["Location", session.location])
    if session.weather:
        info_data.append(["Weather", session.weather])
    if session.started_at:
        info_data.append(["Started", session.started_at.strftime("%Y-%m-%d %H:%M")])
    if session.completed_at:
        info_data.append(["Completed", session.completed_at.strftime("%Y-%m-%d %H:%M")])

    info_table = Table(info_data, colWidths=[1.5 * inch, 4 * inch])
    info_table.setStyle(TableStyle([
        ("FONTNAME", (0, 0), (0, -1), "Helvetica-Bold"),
        ("FONTSIZE", (0, 0), (-1, -1), 10),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 4),
    ]))
    elements.append(info_table)
    elements.append(Spacer(1, 18))

    if session.ends:
        elements.append(Paragraph("Scorecard", styles["Heading2"]))
        elements.append(Spacer(1, 6))

        header = ["End"]
        max_arrows = max(len(end.arrows) for end in session.ends) if session.ends else 0
        for i in range(1, max_arrows + 1):
            header.append(f"A{i}")
        header.append("Total")

        table_data = [header]
        for end in session.ends:
            row = [str(end.end_number)]
            for arrow in end.arrows:
                row.append(arrow.score_value)
            while len(row) < max_arrows + 1:
                row.append("")
            row.append(str(end.end_total))
            table_data.append(row)

        t = Table(table_data)
        t.setStyle(TableStyle([
            ("BACKGROUND", (0, 0), (-1, 0), colors.HexColor("#059669")),
            ("TEXTCOLOR", (0, 0), (-1, 0), colors.white),
            ("FONTNAME", (0, 0), (-1, 0), "Helvetica-Bold"),
            ("FONTSIZE", (0, 0), (-1, -1), 9),
            ("GRID", (0, 0), (-1, -1), 0.5, colors.grey),
            ("ALIGN", (0, 0), (-1, -1), "CENTER"),
            ("BOTTOMPADDING", (0, 0), (-1, -1), 4),
            ("TOPPADDING", (0, 0), (-1, -1), 4),
            ("ROWBACKGROUNDS", (0, 1), (-1, -1), [colors.white, colors.HexColor("#f0fdf4")]),
        ]))
        elements.append(t)

    if session.notes:
        elements.append(Spacer(1, 12))
        elements.append(Paragraph("Notes", styles["Heading3"]))
        elements.append(Paragraph(session.notes, styles["Normal"]))

    doc.build(elements)
    return buffer.getvalue()
