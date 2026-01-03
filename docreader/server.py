"""
Docreader gRPC Server
PDF 파싱 및 청킹 서비스
"""
import logging
import os
import sys
import tempfile
from concurrent import futures

import grpc
from proto import docreader_pb2, docreader_pb2_grpc
from parser.pdf_bbox_extractor import PDFBboxExtractor
from chunker.page_aware_chunker import PageAwareChunker

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class DocReaderServicer(docreader_pb2_grpc.DocReaderServicer):
    """DocReader gRPC Service Implementation"""

    def __init__(self):
        self.bbox_extractor = PDFBboxExtractor()

    def ParsePDF(self, request, context):
        """Parse PDF file and extract chunks with bboxes"""
        try:
            logger.info(f"Parsing PDF: {request.filename}")

            # Save PDF to temp file (pdfplumber needs file path)
            with tempfile.NamedTemporaryFile(delete=False, suffix='.pdf') as tmp_file:
                tmp_file.write(request.file_content)
                tmp_path = tmp_file.name

            try:
                # Get total page count first
                import pdfplumber
                with pdfplumber.open(tmp_path) as pdf:
                    total_pages = len(pdf.pages)

                # Extract paragraphs with bboxes for all pages
                logger.info(f"Extracting bboxes from {total_pages} pages...")
                page_data = self.bbox_extractor.extract_page_bboxes(tmp_path)

                # Chunk with page awareness
                chunk_size = request.chunk_config.chunk_size or 500
                chunk_overlap = request.chunk_config.chunk_overlap or 50

                chunker = PageAwareChunker(chunk_size=chunk_size, chunk_overlap=chunk_overlap)
                chunks = chunker.chunk_pages(page_data)

                logger.info(f"Created {len(chunks)} chunks from {total_pages} pages")

                # Create response chunks
                response_chunks = []

                for chunk in chunks:
                    # Create chunk message with accurate page and bbox
                    chunk_msg = docreader_pb2.Chunk(
                        content=chunk["text"],
                        chunk_index=chunk["chunk_index"],
                        page_number=chunk["page_number"],
                        start_pos=chunk["start_pos"],
                        end_pos=chunk["end_pos"]
                    )

                    # Add bbox if available
                    if chunk["bbox"]:
                        bbox_data = docreader_pb2.BoundingBox(
                            x1=chunk["bbox"]["x1"],
                            y1=chunk["bbox"]["y1"],
                            x2=chunk["bbox"]["x2"],
                            y2=chunk["bbox"]["y2"]
                        )
                        chunk_msg.bbox.CopyFrom(bbox_data)

                    response_chunks.append(chunk_msg)

                # Log bbox statistics
                chunks_with_bbox = sum(1 for c in chunks if c["bbox"] is not None)
                logger.info(f"✓ Successfully parsed PDF: {len(response_chunks)} chunks, "
                           f"{chunks_with_bbox}/{len(chunks)} with bbox ({chunks_with_bbox*100//len(chunks)}%), "
                           f"{total_pages} pages")

                return docreader_pb2.ParseResponse(
                    chunks=response_chunks,
                    total_pages=total_pages
                )

            finally:
                # Clean up temp file
                if os.path.exists(tmp_path):
                    os.remove(tmp_path)

        except Exception as e:
            logger.error(f"Error parsing PDF: {e}", exc_info=True)
            return docreader_pb2.ParseResponse(
                error=str(e)
            )


def serve():
    """Start gRPC server"""
    port = os.environ.get('GRPC_PORT', '50051')

    # Set max message size to 100MB (to handle large PDFs)
    MAX_MESSAGE_LENGTH = 100 * 1024 * 1024  # 100 MB
    options = [
        ('grpc.max_send_message_length', MAX_MESSAGE_LENGTH),
        ('grpc.max_receive_message_length', MAX_MESSAGE_LENGTH),
    ]

    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=10),
        options=options
    )
    docreader_pb2_grpc.add_DocReaderServicer_to_server(DocReaderServicer(), server)

    server.add_insecure_port(f'[::]:{port}')
    server.start()

    logger.info(f"DocReader gRPC server started on port {port} (max message: {MAX_MESSAGE_LENGTH / (1024*1024):.0f}MB)")

    try:
        server.wait_for_termination()
    except KeyboardInterrupt:
        logger.info("Shutting down server...")
        server.stop(0)


if __name__ == '__main__':
    serve()
