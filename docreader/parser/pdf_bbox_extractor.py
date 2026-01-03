"""
PDF Bounding Box Extractor
WeKnora 프로젝트 참조하여 구현
"""
import pdfplumber
from typing import Dict, List, Tuple
import logging

logger = logging.getLogger(__name__)


class PDFBboxExtractor:
    """Extract bounding box information from PDF"""

    def __init__(self):
        self.x_tolerance = 3
        self.y_tolerance = 3

    def extract_page_bboxes(self, pdf_path: str) -> Dict[int, List[Dict]]:
        """
        Extract bounding boxes for each page

        Returns:
            {
                1: [{"text": "...", "bbox": {"x1": 100, "y1": 200, ...}}, ...],
                2: [...],
            }
        """
        page_bboxes = {}

        try:
            with pdfplumber.open(pdf_path) as pdf:
                for page_num, page in enumerate(pdf.pages, start=1):
                    words = page.extract_words(
                        x_tolerance=self.x_tolerance,
                        y_tolerance=self.y_tolerance,
                        keep_blank_chars=False
                    )

                    if not words:
                        logger.warning(f"No text found on page {page_num}")
                        page_bboxes[page_num] = []
                        continue

                    paragraphs = self._group_words_into_paragraphs(words)
                    page_bboxes[page_num] = paragraphs

        except Exception as e:
            logger.error(f"Error extracting bboxes: {e}")
            raise

        return page_bboxes

    def _group_words_into_paragraphs(self, words: List[Dict]) -> List[Dict]:
        """Group words into paragraphs"""
        if not words:
            return []

        paragraphs = []
        current_para = {
            "text": "",
            "bbox": {"x1": float('inf'), "y1": float('inf'), "x2": 0, "y2": 0}
        }

        prev_bottom = None

        for word in words:
            is_new_para = (prev_bottom is not None and
                          word['top'] > prev_bottom + 15)

            if is_new_para and len(current_para["text"]) > 0:
                paragraphs.append(current_para)
                current_para = {
                    "text": word["text"],
                    "bbox": {
                        "x1": word["x0"],
                        "y1": word["top"],
                        "x2": word["x1"],
                        "y2": word["bottom"]
                    }
                }
            else:
                separator = " " if not current_para["text"] else " "
                current_para["text"] += separator + word["text"]
                current_para["bbox"]["x1"] = min(current_para["bbox"]["x1"], word["x0"])
                current_para["bbox"]["y1"] = min(current_para["bbox"]["y1"], word["top"])
                current_para["bbox"]["x2"] = max(current_para["bbox"]["x2"], word["x1"])
                current_para["bbox"]["y2"] = max(current_para["bbox"]["y2"], word["bottom"])

            prev_bottom = word["bottom"]

        if len(current_para["text"]) > 0:
            paragraphs.append(current_para)

        paragraphs = [p for p in paragraphs if len(p["text"]) > 20]

        return paragraphs

    def match_chunk_to_bbox(
        self, chunk_text: str, page_bboxes: List[Dict]
    ) -> Tuple[Dict, float]:
        """Match chunk text to bbox"""
        if not page_bboxes:
            return None, 0.0

        chunk_text_clean = chunk_text.strip()[:100].lower()

        best_bbox = None
        best_score = 0.0

        for para in page_bboxes:
            para_text = para["text"].lower()

            if para_text.startswith(chunk_text_clean):
                return para["bbox"], 1.0

            if chunk_text_clean in para_text:
                overlap = len(chunk_text_clean) / len(para_text)
                if overlap > best_score:
                    best_bbox = para["bbox"]
                    best_score = overlap

        if best_score > 0.5:
            return best_bbox, best_score

        return None, 0.0

    def estimate_chunk_page(
        self, chunk_index: int, total_chunks: int, total_pages: int
    ) -> int:
        """Estimate page number for a chunk"""
        if total_pages == 0 or total_chunks == 0:
            return 1

        estimated_page = int((chunk_index / total_chunks) * total_pages) + 1
        return max(1, min(estimated_page, total_pages))
