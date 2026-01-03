"""
Text Chunking with overlap
WeKnora의 TextSplitter 참조
"""
from typing import List, Tuple


class TextChunker:
    """Split text into chunks with overlap"""

    def __init__(self, chunk_size: int = 500, chunk_overlap: int = 50):
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap

    def chunk_text(self, text: str) -> List[Tuple[int, int, str]]:
        """
        Split text into chunks

        Returns:
            List of (start_pos, end_pos, chunk_text)
        """
        if not text:
            return []

        chunks = []
        start = 0
        text_len = len(text)

        while start < text_len:
            end = min(start + self.chunk_size, text_len)

            # Try to break at sentence boundary
            if end < text_len:
                # Look for sentence endings
                for sep in ['. ', '.\n', '! ', '?\n', '? ']:
                    last_sep = text.rfind(sep, start, end)
                    if last_sep != -1:
                        end = last_sep + len(sep)
                        break

            chunk_text = text[start:end].strip()
            if chunk_text:
                chunks.append((start, end, chunk_text))

            # Move start position with overlap
            start = end - self.chunk_overlap if end - self.chunk_overlap > start else end

        return chunks
