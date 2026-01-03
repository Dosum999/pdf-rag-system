"""
Page-Aware Chunking with Bounding Box Tracking
페이지 정보와 bbox를 유지하면서 청킹
"""
from typing import List, Dict, Optional
import logging

logger = logging.getLogger(__name__)


class PageAwareChunker:
    """페이지 정보를 유지하면서 텍스트를 청킹"""

    def __init__(self, chunk_size: int = 500, chunk_overlap: int = 50):
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap

    def chunk_pages(self, page_data: Dict[int, List[Dict]]) -> List[Dict]:
        """
        페이지별 paragraph 데이터를 청킹

        Args:
            page_data: {
                1: [{"text": "...", "bbox": {"x1": ..., "y1": ..., "x2": ..., "y2": ...}}, ...],
                2: [...],
            }

        Returns:
            List of chunks with page and bbox info:
            [
                {
                    "text": "chunk content",
                    "page_number": 1,
                    "bbox": {"x1": ..., "y1": ..., "x2": ..., "y2": ...},
                    "start_pos": 0,
                    "end_pos": 500
                },
                ...
            ]
        """
        chunks = []
        global_pos = 0  # Track position across all pages

        for page_num in sorted(page_data.keys()):
            paragraphs = page_data[page_num]

            if not paragraphs:
                continue

            # Chunk within this page
            page_chunks = self._chunk_paragraphs(
                paragraphs=paragraphs,
                page_number=page_num,
                start_global_pos=global_pos
            )

            chunks.extend(page_chunks)

            # Update global position
            total_page_text = " ".join([p["text"] for p in paragraphs])
            global_pos += len(total_page_text) + 1  # +1 for page break

        return chunks

    def _chunk_paragraphs(
        self,
        paragraphs: List[Dict],
        page_number: int,
        start_global_pos: int
    ) -> List[Dict]:
        """페이지 내의 문단들을 청킹"""
        chunks = []

        # Build continuous text from paragraphs
        para_texts = []
        para_positions = []  # Track which paragraph each character belongs to
        current_pos = 0

        for idx, para in enumerate(paragraphs):
            text = para["text"]
            para_texts.append(text)

            # Record paragraph index for each character range
            para_positions.append({
                "para_idx": idx,
                "start": current_pos,
                "end": current_pos + len(text),
                "bbox": para["bbox"]
            })

            current_pos += len(text) + 1  # +1 for space

        full_text = " ".join(para_texts)

        # Create chunks with character-based splitting
        start = 0
        chunk_index = 0

        while start < len(full_text):
            end = min(start + self.chunk_size, len(full_text))

            # Try to break at sentence boundary
            if end < len(full_text):
                for sep in ['. ', '.\n', '! ', '!\n', '? ', '?\n', '。 ']:
                    last_sep = full_text.rfind(sep, start, end)
                    if last_sep != -1:
                        end = last_sep + len(sep)
                        break

            chunk_text = full_text[start:end].strip()

            if chunk_text:
                # Find which paragraphs this chunk spans
                chunk_bbox = self._get_chunk_bbox(
                    start_pos=start,
                    end_pos=end,
                    para_positions=para_positions
                )

                chunks.append({
                    "text": chunk_text,
                    "page_number": page_number,
                    "bbox": chunk_bbox,
                    "start_pos": start_global_pos + start,
                    "end_pos": start_global_pos + end,
                    "chunk_index": chunk_index
                })

                chunk_index += 1

            # Move with overlap
            start = end - self.chunk_overlap if end - self.chunk_overlap > start else end

        return chunks

    def _get_chunk_bbox(
        self,
        start_pos: int,
        end_pos: int,
        para_positions: List[Dict]
    ) -> Optional[Dict]:
        """
        청크가 걸쳐있는 문단들의 bbox를 병합

        Returns:
            Merged bounding box or None if no overlap
        """
        overlapping_paras = []

        for para_info in para_positions:
            # Check if chunk overlaps with this paragraph
            if not (end_pos <= para_info["start"] or start_pos >= para_info["end"]):
                overlapping_paras.append(para_info)

        if not overlapping_paras:
            return None

        # Merge bboxes of overlapping paragraphs
        merged_bbox = {
            "x1": min(p["bbox"]["x1"] for p in overlapping_paras),
            "y1": min(p["bbox"]["y1"] for p in overlapping_paras),
            "x2": max(p["bbox"]["x2"] for p in overlapping_paras),
            "y2": max(p["bbox"]["y2"] for p in overlapping_paras)
        }

        return merged_bbox
