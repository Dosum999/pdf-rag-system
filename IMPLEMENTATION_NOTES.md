# Implementation Notes

## 완성된 기능

✅ **프로젝트 구조**
- Docker Compose 기반 multi-container 설정
- Backend (Go), Docreader (Python), Frontend (Vue 3) 분리
- PostgreSQL + pgvector 벡터 DB

✅ **Docreader (Python gRPC 서비스)**
- PDF 파싱 (pdfplumber)
- Bbox 추출 및 매칭
- Text chunking with overlap
- gRPC 서버 구현 완료

✅ **Database Schema**
- Documents 테이블
- Chunks 테이블 with vector(1536) embedding
- Bounding box 좌표 (bbox_x1, y1, x2, y2)
- Indexes for performance

✅ **Backend 구조 (Go)**
- Domain models (Document, Chunk)
- Config, Database initialization
- Project structure 완료

✅ **Frontend 구조 (Vue 3)**
- Main layout with document list
- Chat/Query interface
- PDF Viewer component (pdf.js)
- Citation Card component
- API integration

## 구현이 필요한 부분

다음 기능들은 구조와 가이드만 제공되어 있으며 실제 코드 구현이 필요합니다:

### 1. Backend Service Layer
**파일**: `backend/internal/service/`

필요한 구현:
- `document.go`: PDF 업로드 처리, gRPC 호출, 임베딩 생성
- `chat.go`: RAG 파이프라인, 벡터 검색, LLM 호출
- `embedding.go`: OpenAI Embedding API 호출

참조: `backend/README.md`에 핵심 로직 스니펫 포함

### 2. Backend Repository Layer
**파일**: `backend/internal/repository/`

필요한 구현:
- `document.go`: Document CRUD
- `chunk.go`: Chunk CRUD + pgvector 유사도 검색

pgvector 쿼리 예시:
```go
SELECT *, 1 - (embedding <=> $1) as score
FROM chunks
ORDER BY embedding <=> $1
LIMIT $2
```

### 3. Backend API Handlers
**파일**: `backend/internal/api/`

필요한 구현:
- `document.go`: Upload, List, Get, Delete handlers
- `chat.go`: Query handler

### 4. Backend gRPC Client
**파일**: `backend/internal/client/`

필요한 구현:
- `docreader.go`: Docreader gRPC 클라이언트
- `llm.go`: OpenAI API 클라이언트

### 5. Frontend 추가 설정
**파일**: `frontend/`

필요한 구현:
- `vite.config.ts`: Vite 설정
- `src/main.ts`: Vue app 진입점
- `src/router/index.ts`: Vue Router 설정
- `nginx.conf`: Production용 Nginx 설정

## 구현 우선순위

### Phase 1: 핵심 기능 (필수)
1. **Backend Repository** - DB 접근 로직
2. **Backend Service** - 비즈니스 로직
3. **Backend API Handlers** - HTTP 엔드포인트
4. **Backend gRPC Client** - Docreader 연동

### Phase 2: 통합
5. **Embedding Service** - 벡터 생성
6. **LLM Client** - RAG 답변 생성
7. **Frontend 설정 파일** - Vue app 실행 가능하게

### Phase 3: 테스트 및 개선
8. 전체 워크플로우 테스트
9. 성능 최적화
10. 에러 핸들링 강화

## 빠른 시작을 위한 최소 구현

다음 파일들만 구현하면 기본 동작 가능:

1. `backend/internal/repository/chunk.go`
   - VectorSearch() 메서드

2. `backend/internal/service/document.go`
   - Upload() 메서드

3. `backend/internal/service/chat.go`
   - Query() 메서드

4. `backend/internal/api/document.go`
   - Upload handler

5. `backend/internal/api/chat.go`
   - Query handler

6. `backend/internal/client/docreader.go`
   - gRPC 연결

나머지는 mock 데이터나 간단한 구현으로 대체 가능합니다.

## 참고사항

- **WeKnora 참조**: 실제 구현 시 `WeKnora-main` 프로젝트의 해당 파일들을 참조
- **API 키**: `.env` 파일에 OpenAI API 키 필수
- **Protobuf**: docreader.proto 수정 시 양쪽 모두 재생성 필요
- **벡터 차원**: OpenAI text-embedding-3-small 사용 시 1536 차원

## 추가 개선 아이디어

1. **인증/권한**: JWT 기반 사용자 인증
2. **파일 스토리지**: S3/MinIO로 PDF 파일 저장
3. **캐싱**: Redis로 검색 결과 캐싱
4. **모니터링**: Prometheus + Grafana
5. **로깅**: Structured logging (zerolog)
6. **테스트**: Unit tests, Integration tests
