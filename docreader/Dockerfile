FROM python:3.10-slim

WORKDIR /app

# Install system dependencies (netcat for healthcheck)
RUN apt-get update && apt-get install -y netcat-traditional && rm -rf /var/lib/apt/lists/*

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

# Remove any old generated protobuf files and regenerate
RUN rm -f proto/*_pb2.py proto/*_pb2_grpc.py && \
    python -m grpc_tools.protoc \
    -I./proto \
    --python_out=./proto \
    --grpc_python_out=./proto \
    proto/docreader.proto && \
    sed -i 's/^import docreader_pb2/from . import docreader_pb2/' proto/docreader_pb2_grpc.py

# Expose gRPC port
EXPOSE 50051

# Run server
CMD ["python", "server.py"]
