# Backend - PDF RAG System

Go 백엔드 API 서버입니다.

## 주요 기능

- PDF 파일 업로드 및 처리
- 벡터 검색
- RAG 기반 질의응답
- Citation 정보 제공

## 구조

```
backend/
├── cmd/server/main.go          # 진입점
├── internal/
│   ├── api/                    # HTTP handlers
│   │   ├── document.go         # 문서 업로드/조회
│   │   └── chat.go             # 질의응답
│   ├── service/                # 비즈니스 로직
│   │   ├── document.go         # 문서 처리 서비스
│   │   ├── chat.go             # RAG 서비스
│   │   └── embedding.go        # 임베딩 서비스
│   ├── repository/             # 데이터 액세스
│   │   ├── document.go         # 문서 CRUD
│   │   └── chunk.go            # 청크 CRUD + 벡터 검색
│   ├── client/                 # 외부 서비스 클라이언트
│   │   ├── docreader.go        # gRPC 클라이언트
│   │   └── llm.go              # LLM API 클라이언트
│   └── domain/                 # 도메인 모델
│       ├── document.go
│       └── chunk.go
└── pkg/
    ├── config/                 # 설정
    └── database/               # DB 초기화
```

## API 엔드포인트

### 문서 관리

**PDF 업로드**
```
POST /api/v1/documents/upload
Content-Type: multipart/form-data

Response:
{
  "id": "doc-uuid",
  "filename": "sample.pdf",
  "status": "processing"
}
```

**문서 목록**
```
GET /api/v1/documents

Response:
{
  "documents": [...]
}
```

### 질의응답

**Query**
```
POST /api/v1/chat/query
{
  "query": "질문 내용",
  "document_ids": ["doc-1", "doc-2"]
}

Response:
{
  "answer": "답변...",
  "citations": [
    {
      "document_id": "...",
      "filename": "sample.pdf",
      "page_number": 5,
      "content": "발췌 내용...",
      "bbox": {"x1": 72.5, "y1": 150, "x2": 520, "y2": 180},
      "score": 0.85
    }
  ]
}
```

## 구현 세부사항

### 1. Document Service (internal/service/document.go)

```go
// 문서 업로드 처리
func (s *DocumentService) Upload(file multipart.File, filename string) (*domain.Document, error) {
    // 1. 파일 저장
    // 2. gRPC로 docreader 호출하여 PDF 파싱
    // 3. chunks 저장
    // 4. 각 chunk에 대해 embedding 생성
    // 5. pgvector에 저장
}
```

### 2. Chat Service (internal/service/chat.go)

```go
// RAG 파이프라인
func (s *ChatService) Query(query string, docIDs []string) (*ChatResponse, error) {
    // 1. 쿼리 임베딩 생성
    // 2. 벡터 유사도 검색 (pgvector)
    // 3. 상위 K개 chunks 가져오기
    // 4. LLM에 컨텍스트와 함께 전달
    // 5. 답변 + citations 반환
}
```

### 3. Chunk Repository (internal/repository/chunk.go)

```go
// 벡터 유사도 검색
func (r *ChunkRepository) VectorSearch(embedding []float64, topK int) ([]domain.SearchResult, error) {
    query := `
        SELECT c.*, d.filename,
               1 - (c.embedding <=> $1) as score
        FROM chunks c
        JOIN documents d ON c.document_id = d.id
        WHERE d.id = ANY($2)
        ORDER BY c.embedding <=> $1
        LIMIT $3
    `
    // pgvector cosine similarity search
}
```

## 실행

```bash
# 의존성 설치
go mod download

# 실행
go run cmd/server/main.go

# 빌드
go build -o pdf-rag-server cmd/server/main.go
```

## 환경 변수

`.env` 파일 참조
