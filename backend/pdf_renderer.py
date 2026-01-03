"""
PDF page rendering to image with bbox highlighting
"""
import io
from pdf2image import convert_from_path
from PIL import Image, ImageDraw


def render_pdf_page_to_image(pdf_path: str, page_number: int, bbox=None, dpi=150):
    """
    Render a PDF page to PNG image with optional bbox highlighting

    Args:
        pdf_path: Path to PDF file
        page_number: Page number (1-indexed)
        bbox: Optional dict with x1, y1, x2, y2 coordinates
        dpi: Resolution for rendering (default 150)

    Returns:
        bytes: PNG image data
    """
    # Convert PDF page to image
    images = convert_from_path(
        pdf_path,
        dpi=dpi,
        first_page=page_number,
        last_page=page_number
    )

    if not images:
        raise ValueError(f"Failed to render page {page_number}")

    img = images[0]

    # Draw bbox if provided
    if bbox:
        draw = ImageDraw.Draw(img)

        # Scale bbox coordinates to image size
        zoom = dpi / 72  # 72 DPI is default PDF resolution
        x1 = int(bbox['x1'] * zoom)
        y1 = int(bbox['y1'] * zoom)
        x2 = int(bbox['x2'] * zoom)
        y2 = int(bbox['y2'] * zoom)

        # Draw rectangle with border
        border_width = max(2, int(3 * zoom))
        draw.rectangle(
            [x1, y1, x2, y2],
            outline='#0052d9',  # TDesign primary color
            width=border_width
        )

        # Draw semi-transparent fill
        overlay = Image.new('RGBA', img.size, (0, 0, 0, 0))
        overlay_draw = ImageDraw.Draw(overlay)
        overlay_draw.rectangle(
            [x1, y1, x2, y2],
            fill=(0, 82, 217, 30)  # RGBA with 30 alpha
        )
        img = img.convert('RGBA')
        img = Image.alpha_composite(img, overlay)
        img = img.convert('RGB')

    # Convert to bytes
    img_bytes = io.BytesIO()
    img.save(img_bytes, format='PNG', optimize=True)
    img_bytes.seek(0)

    return img_bytes.getvalue()


def get_pdf_page_count(pdf_path: str) -> int:
    """Get total number of pages in PDF"""
    from pdf2image import pdfinfo_from_path
    info = pdfinfo_from_path(pdf_path)
    return info.get("Pages", 0)
