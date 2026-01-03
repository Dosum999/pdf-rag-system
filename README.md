# PDF RAG System with Citation

PDF 문서 기반 질의응답 시스템 - 출처 추적 및 하이라이팅 기능 제공

## 프로젝트 개요

여러 개의 PDF 문서를 업로드하고, 자연어 질문에 대해 답변을 생성하며, **답변의 근거가 되는 PDF 파일명, 페이지 번호, 문단 위치를 함께 제시**하는 RAG(Retrieval-Augmented Generation) 시스템입니다.

### 주요 기능

- **PDF 업로드 및 처리**
  - 복수 PDF 파일 업로드 지원 (최대 200MB)
  - PDF 텍스트 및 좌표(bbox) 자동 추출
  - 페이지별 청킹 및 벡터화

- **질의응답 with 출처 표기**
  - 자연어 질의 처리
  - 답변과 함께 출처 정보 제공
    - PDF 파일명
    - 페이지 번호
    - 근거 텍스트 발췌
    - Bounding box 좌표
  - 문서에 근거가 없는 경우 "문서에서 확인 불가" 명시

- **PDF 뷰어 with 하이라이팅**
  - PDF.js 기반 인터랙티브 뷰어
  - 페이지 탐색 및 줌 기능
  - 근거 문단 자동 하이라이팅
  - 좌표 기반 정확한 위치 표시

- **성능**
  - 질의응답 처리 시간 < 10초
  - 벡터 검색 기반 빠른 검색
  - 로컬 LLM 사용으로 비용 절감

## 기술 스택

### 백엔드
- **Go 1.21+** - REST API 서버
- **Gin** - HTTP 프레임워크
- **GORM** - ORM
- **gRPC** - 마이크로서비스 통신

### 문서 처리 (Docreader)
- **Python 3.10+** - 문서 파싱 서버
- **pdfplumber** - PDF 텍스트 및 좌표 추출
- **PyMuPDF** - PDF 렌더링
- **gRPC** - 서비스 통신

### 프론트엔드
- **Vue 3** - UI 프레임워크
- **TypeScript** - 타입 안전성
- **Vite** - 빌드 도구
- **TDesign** - UI 컴포넌트 라이브러리
- **PDF.js** - PDF 뷰어

### 인프라
- **PostgreSQL 15+** - 메인 데이터베이스
- **pgvector** - 벡터 검색 확장
- **Docker & Docker Compose** - 컨테이너화
- **Ollama** - 로컬 LLM 서버 (무료)

### AI 모델
- **LLM**: Llama 3.1 (via Ollama)
- **Embedding**: nomic-embed-text (via Ollama)

## 시스템 아키텍처

```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐
│   Frontend  │─────▶│   Backend    │─────▶│  Docreader  │
│   (Vue 3)   │      │     (Go)     │      │  (Python)   │
│   :3000     │      │    :8080     │      │   :50051    │
└─────────────┘      └──────┬───────┘      └─────────────┘
                            │
                     ┌──────▼───────┐      ┌─────────────┐
                     │ PostgreSQL   │      │   Ollama    │
                     │  + pgvector  │      │   (LLM)     │
                     │    :5432     │      │   :11434    │
                     └──────────────┘      └─────────────┘
```

## 프로젝트 구조

```
pdf-rag-system/
├── backend/              # Go 백엔드 API 서버
│   ├── cmd/server/       # 메인 엔트리포인트
│   ├── internal/
│   │   ├── api/          # HTTP 핸들러
│   │   ├── domain/       # 도메인 모델
│   │   ├── service/      # 비즈니스 로직
│   │   ├── repository/   # 데이터 액세스
│   │   └── client/       # gRPC 클라이언트
│   ├── pkg/config/       # 설정 관리
│   └── Dockerfile
│
├── docreader/            # Python 문서 처리 서버
│   ├── proto/            # Protobuf 정의
│   ├── parser/           # PDF 파서 (bbox 추출)
│   ├── chunker/          # 텍스트 청킹
│   ├── server.py         # gRPC 서버
│   ├── requirements.txt
│   └── Dockerfile
│
├── frontend/             # Vue 3 프론트엔드
│   ├── src/
│   │   ├── components/   # UI 컴포넌트
│   │   ├── views/        # 페이지 뷰
│   │   ├── api/          # API 클라이언트
│   │   └── main.ts
│   ├── package.json
│   └── Dockerfile
│
├── database/
│   └── migrations/       # DB 마이그레이션
│
├── docker-compose.yml    # 전체 스택 오케스트레이션
├── .env                  # 환경 변수
└── README.md
```

## 설치 및 실행

### 사전 요구사항

필수:
- **Docker Desktop** (최신 버전)
- **Docker Compose** v2.0+
- uploads 하위 폴더 생성

선택 (로컬 개발 시):
- Go 1.21+
- Python 3.10+
- Node.js 18+

### 1. Ollama 설치 및 모델 다운로드

이 프로젝트는 **완전 무료 로컬 LLM**을 사용합니다. OpenAI API 키가 필요 없습니다.

#### Windows

1. [Ollama 공식 사이트](https://ollama.com/download)에서 설치 프로그램 다운로드
2. 설치 후 PowerShell 또는 CMD에서 모델 다운로드:

```bash
# LLM 모델 다운로드 (약 4.7GB)
ollama pull llama3.1

# Embedding 모델 다운로드 (약 274MB)
ollama pull nomic-embed-text

# 설치 확인
ollama list
```

#### macOS

```bash
# Homebrew로 설치
brew install ollama

# 모델 다운로드
ollama pull llama3.1
ollama pull nomic-embed-text

# 설치 확인
ollama list
```

#### Linux

```bash
# 설치
curl -fsSL https://ollama.com/install.sh | sh

# 모델 다운로드
ollama pull llama3.1
ollama pull nomic-embed-text

# 설치 확인
ollama list
```

**참고**: 첫 실행 시 Ollama가 자동으로 백그라운드 서비스로 실행됩니다 (포트 11434).

### 2. 프로젝트 클론 및 환경 설정

```bash
# 프로젝트 디렉토리로 이동
cd "AI work/pdf-rag-system"

# 환경 변수 확인 (.env.example파일을 .env로 변경)
cat .env

# .env 파일을 필요에 따라 수정
# LLM_MODEL, EMBEDDING_MODEL 등은 이미 설정되어 있음
```

### 3. Docker Compose로 전체 스택 실행 (권장)

```bash
# 모든 서비스 빌드 및 실행
docker-compose up --build -d

# 로그 확인
docker-compose logs -f

# 특정 서비스 로그만 확인
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f docreader
```

**서비스 시작 순서**:
1. PostgreSQL 시작 및 초기화 (약 10초)
2. Docreader gRPC 서버 시작 (약 5초)
3. Backend API 서버 시작 (약 5초)
4. Frontend 시작 (약 3초)

전체 스택이 완전히 시작되기까지 약 20-30초 소요됩니다.

### 4. 애플리케이션 접속

서비스가 모두 시작되면:

- **프론트엔드**: http://localhost:3000
- **백엔드 API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

### 5. 서비스 중지

```bash
# 모든 서비스 중지
docker-compose down

# 볼륨까지 삭제 (데이터베이스 초기화)
docker-compose down -v
```

## 로컬 개발 모드

Docker를 사용하지 않고 로컬에서 개발하려면:

### PostgreSQL 준비

```bash
# Docker로 PostgreSQL + pgvector 실행
docker run -d \
  --name pdf-rag-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=pdf_rag_db \
  -p 5432:5432 \
  ankane/pgvector:latest

# 마이그레이션 실행
docker exec -i pdf-rag-postgres psql -U postgres -d pdf_rag_db < database/migrations/001_init.sql
```

### Docreader 서버 실행

```bash
cd docreader

# Python 가상환경 생성 (선택)
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# 의존성 설치
pip install -r requirements.txt

# Protobuf 컴파일
python -m grpc_tools.protoc \
  -I./proto \
  --python_out=./proto \
  --grpc_python_out=./proto \
  proto/docreader.proto

# 서버 실행 (포트 50051)
python server.py
```

### Backend 서버 실행

```bash
cd backend

# 의존성 다운로드
go mod download

# 환경 변수 설정 (.env 파일 사용)
export $(cat ../.env | xargs)

# 서버 실행 (포트 8080)
go run cmd/server/main.go
```

### Frontend 개발 서버 실행

```bash
cd frontend

# 의존성 설치
npm install

# 개발 서버 실행 (포트 5173, Vite 기본 포트)
npm run dev

# 프로덕션 빌드
npm run build
```

## 사용 방법

### 1. PDF 업로드

1. http://localhost:3000 접속
2. "PDF 업로드" 버튼 클릭
3. PDF 파일 선택 (최대 50MB)
4. 업로드 완료 대기 (진행률 표시)

### 2. 질문하기

1. 업로드된 문서 선택
2. 질문 입력창에 질문 작성
3. "질문하기" 버튼 클릭
4. 답변 및 출처 확인

### 3. 출처 확인

1. 답변 하단의 "출처" 카드 확인
2. 각 출처에는 다음 정보 포함:
   - 파일명
   - 페이지 번호
   - 발췌 내용
   - 신뢰도 점수
3. "원문 보기" 클릭 시 PDF 뷰어 열림
4. 해당 페이지로 자동 이동 및 하이라이팅

## API 문서

### PDF 업로드

```http
POST /api/documents/upload
Content-Type: multipart/form-data

file: <PDF 파일>
```

**응답**:
```json
{
  "document_id": "doc-uuid-123",
  "filename": "sample.pdf",
  "pages": 25,
  "chunks_created": 150
}
```

### 질의응답

```http
POST /api/chat/query
Content-Type: application/json

{
  "query": "이 문서의 주요 내용은 무엇인가요?",
  "document_ids": ["doc-uuid-123"]
}
```

**응답**:
```json
{
  "answer": "이 문서는 PDF 기반 RAG 시스템에 대해 설명합니다...",
  "citations": [
    {
      "document_id": "doc-uuid-123",
      "filename": "sample.pdf",
      "page_number": 5,
      "content": "RAG 시스템은 검색 증강 생성을 의미하며...",
      "bbox": {
        "x1": 72.5,
        "y1": 150.2,
        "x2": 520.3,
        "y2": 180.7
      },
      "score": 0.85
    }
  ],
  "processing_time_ms": 2340
}
```

### Health Check

```http
GET /health
```

**응답**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## 환경 변수 설명

`.env` 파일의 주요 설정:

```bash
# Database
DB_HOST=postgres              # Docker: postgres, 로컬: localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=pdf_rag_db

# LLM (Ollama)
LLM_API_BASE_URL=http://host.docker.internal:11434/v1
LLM_API_KEY=ollama           # Ollama는 키 불필요
LLM_MODEL=llama3.1           # 다운로드한 모델명

# Embedding (Ollama)
EMBEDDING_API_URL=http://host.docker.internal:11434/v1
EMBEDDING_API_KEY=ollama
EMBEDDING_MODEL=nomic-embed-text

# Storage
UPLOAD_DIR=./uploads
MAX_FILE_SIZE=50MB

# Vector Search
VECTOR_DIMENSION=768         # nomic-embed-text 차원
SEARCH_TOP_K=5              # 검색 결과 개수

# Chunking
CHUNK_SIZE=500              # 청크 크기 (토큰)
CHUNK_OVERLAP=50            # 청크 오버랩
```

## 트러블슈팅

### Ollama 연결 오류

**증상**: `connection refused to localhost:11434`

**해결**:
```bash
# Ollama 서비스 상태 확인
ollama list

# 서비스 재시작 (Windows)
# 작업 관리자에서 ollama 프로세스 종료 후 재실행

# 서비스 재시작 (macOS/Linux)
killall ollama
ollama serve
```

### Docker 컨테이너 시작 실패

**증상**: `backend` 서비스가 시작되지 않음

**해결**:
```bash
# 로그 확인
docker-compose logs backend

# 의존성 서비스 확인
docker-compose ps

# PostgreSQL이 ready 상태인지 확인
docker-compose exec postgres pg_isready -U postgres

# 모든 서비스 재시작
docker-compose restart
```

### PDF 업로드 실패

**증상**: PDF 업로드 시 오류 발생

**해결**:
1. 파일 크기 확인 (50MB 이하)
2. PDF 파일 손상 여부 확인
3. Docreader 서비스 로그 확인:
   ```bash
   docker-compose logs docreader
   ```

### 느린 응답 시간

**증상**: 질의응답이 10초 이상 소요

**해결**:
1. Ollama 모델이 올바르게 로드되었는지 확인:
   ```bash
   ollama list
   ```
2. 시스템 리소스 확인 (최소 8GB RAM 권장)
3. `.env`에서 `SEARCH_TOP_K` 값을 줄여보기 (5 → 3)
4. 청크 크기 조정 (`CHUNK_SIZE=500` → `300`)

### pgvector 확장 오류

**증상**: `extension "vector" does not exist`

**해결**:
```bash
# PostgreSQL 컨테이너 재생성
docker-compose down -v
docker-compose up -d postgres

# pgvector 확장 설치 확인
docker-compose exec postgres psql -U postgres -d pdf_rag_db -c "CREATE EXTENSION IF NOT EXISTS vector;"
```

### 프론트엔드 빌드 오류

**증상**: `Cannot find module 'pdfjs-dist'`

**해결**:
```bash
cd frontend

# node_modules 삭제 후 재설치
rm -rf node_modules package-lock.json
npm install

# 또는 Docker 빌드 캐시 삭제
docker-compose build --no-cache frontend
```

## 성능 최적화

### 권장 시스템 사양

- **최소**: CPU 4코어, RAM 8GB, 디스크 20GB
- **권장**: CPU 8코어, RAM 16GB, SSD 50GB

### 성능 튜닝

1. **벡터 검색 최적화**
   - `SEARCH_TOP_K` 조정 (기본값: 5)
   - pgvector 인덱스 재구성 (주기적)

2. **청킹 전략**
   - `CHUNK_SIZE` 조정 (300-1000)
   - `CHUNK_OVERLAP` 조정 (10-100)

3. **LLM 응답 속도**
   - GPU 지원 Ollama 설치 (NVIDIA GPU 사용 시)
   - 더 작은 모델 사용 (llama3.1 → phi3)

## 개발 가이드

### Protobuf 재생성

Protobuf 스키마 수정 시:

```bash
# Python (Docreader)
cd docreader
python -m grpc_tools.protoc \
  -I./proto \
  --python_out=./proto \
  --grpc_python_out=./proto \
  proto/docreader.proto

# Go (Backend)
cd backend/pkg/proto
protoc --go_out=. --go-grpc_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_opt=paths=source_relative \
  ../../docreader/proto/docreader.proto
```

### 데이터베이스 마이그레이션

새로운 마이그레이션 추가:

```bash
# 마이그레이션 파일 생성
touch database/migrations/002_add_new_table.sql

# 마이그레이션 실행 (Docker)
docker-compose exec postgres psql -U postgres -d pdf_rag_db -f /docker-entrypoint-initdb.d/002_add_new_table.sql
```

### 테스트

```bash
# Backend 테스트
cd backend
go test ./...

# Frontend 테스트
cd frontend
npm run test

# E2E 테스트 (수동)
# 1. PDF 업로드
# 2. 질의응답 수행
# 3. 출처 확인
```

## 배포

### Docker 이미지 빌드

```bash
# 전체 스택 빌드
docker-compose build

# 개별 서비스 빌드
docker-compose build backend
docker-compose build frontend
docker-compose build docreader
```

### 프로덕션 환경 변수

프로덕션 배포 시 `.env` 수정:

```bash
# 보안 강화
DB_PASSWORD=강력한_비밀번호

# 로깅 레벨
LOG_LEVEL=info

# CORS 설정
CORS_ALLOWED_ORIGINS=https://yourdomain.com
```

## 라이선스

MIT License

## 참고 문서

- [AIRequirements.md](../AIRequirements.md) - 프로젝트 요구사항 명세
- [IMPLEMENTATION_SUMMARY.md](../IMPLEMENTATION_SUMMARY.md) - 구현 요약
- [MODIFICATION_PLAN.md](../MODIFICATION_PLAN.md) - 수정 계획
- [Ollama 공식 문서](https://ollama.com/docs)
- [pgvector 문서](https://github.com/pgvector/pgvector)

## 지원

문제가 발생하거나 질문이 있으시면 이슈를 등록해주세요.

---

**개발 시간**: 약 20시간
**난이도**: 중상
**주요 기술**: RAG, Vector Search, gRPC, Docker, Vue 3, Go, Python
