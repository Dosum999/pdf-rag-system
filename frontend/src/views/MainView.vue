<template>
  <div class="main-container">
    <t-layout>
      <!-- Header -->
      <t-header class="header">
        <h1>PDF 문서 질의응답 시스템</h1>
      </t-header>

      <t-layout>
        <!-- Sidebar - Document List -->
        <t-aside width="300px" class="sidebar">
          <div class="sidebar-content">
            <h3>문서 목록</h3>
            <t-button theme="primary" block @click="showUploadDialog = true">
              <template #icon><UploadIcon /></template>
              PDF 업로드
            </t-button>

            <div class="document-list">
              <t-list :split="true">
                <t-list-item
                  v-for="doc in documents"
                  :key="doc.id"
                  :class="{ active: selectedDocIds.includes(doc.id) }"
                  @click="toggleDocument(doc.id)"
                >
                  <t-checkbox
                    :value="selectedDocIds.includes(doc.id)"
                    @change="toggleDocument(doc.id)"
                  />
                  <span class="doc-name">{{ doc.filename }}</span>
                  <t-tag size="small">{{ doc.total_pages }}쪽</t-tag>
                </t-list-item>
              </t-list>
            </div>
          </div>
        </t-aside>

        <!-- Main Content -->
        <t-content class="main-content">
          <t-row :gutter="16">
            <!-- Chat/Search Panel -->
            <t-col :span="6">
              <t-card title="질문하기" class="chat-panel">
                <t-textarea
                  v-model="query"
                  placeholder="질문을 입력하세요..."
                  :autosize="{ minRows: 3, maxRows: 6 }"
                />
                <t-button
                  theme="primary"
                  block
                  @click="handleQuery"
                  :loading="loading"
                  :disabled="selectedDocIds.length === 0"
                  style="margin-top: 12px"
                >
                  <template #icon><SearchIcon /></template>
                  검색
                </t-button>

                <!-- Answer -->
                <div v-if="answer" class="answer-section">
                  <h4>답변:</h4>
                  <p>{{ answer }}</p>
                </div>

                <!-- Citations -->
                <div v-if="citations.length > 0" class="citations-section">
                  <h4>출처 ({{ citations.length }}개):</h4>
                  <CitationCard
                    v-for="(citation, idx) in citations"
                    :key="idx"
                    :citation="citation"
                    @view-source="handleViewSource"
                  />
                </div>
              </t-card>
            </t-col>

            <!-- PDF Viewer Panel -->
            <t-col :span="10">
              <t-card title="PDF 뷰어" class="viewer-panel">
                <PDFViewer
                  v-if="viewerConfig"
                  :document-id="viewerConfig.documentId"
                  :page-number="viewerConfig.pageNumber"
                  :bbox="viewerConfig.bbox"
                  :total-pages="viewerConfig.totalPages"
                />
                <t-empty v-else description="출처의 '원문 보기' 버튼을 클릭하면 PDF를 볼 수 있습니다" />
              </t-card>
            </t-col>
          </t-row>
        </t-content>
      </t-layout>
    </t-layout>

    <!-- Upload Dialog -->
    <t-dialog v-model:visible="showUploadDialog" header="PDF 업로드" width="500px">
      <t-upload
        v-model="uploadFiles"
        :auto-upload="false"
        accept=".pdf"
        theme="file"
        :max="1"
        tips="PDF 파일만 지원됩니다"
      />
      <template #footer>
        <t-button @click="showUploadDialog = false">취소</t-button>
        <t-button theme="primary" @click="handleUpload" :disabled="uploadFiles.length === 0">업로드</t-button>
      </template>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { MessagePlugin } from 'tdesign-vue-next'
import { UploadIcon, SearchIcon } from 'tdesign-icons-vue-next'
import PDFViewer from '../components/PDFViewerImage.vue'
import CitationCard from '../components/CitationCard.vue'
import api from '../api'

const documents = ref([])
const selectedDocIds = ref([])
const query = ref('')
const answer = ref('')
const citations = ref([])
const loading = ref(false)
const viewerConfig = ref(null)
const showUploadDialog = ref(false)
const uploadFiles = ref([])

const loadDocuments = async () => {
  try {
    const response = await api.getDocuments()
    documents.value = response.data.documents
  } catch (error) {
    MessagePlugin.error('Failed to load documents')
  }
}

const toggleDocument = (docId) => {
  const index = selectedDocIds.value.indexOf(docId)
  if (index > -1) {
    selectedDocIds.value.splice(index, 1)
  } else {
    selectedDocIds.value.push(docId)
  }
}

const handleQuery = async () => {
  if (!query.value.trim()) {
    MessagePlugin.warning('Please enter a question')
    return
  }

  loading.value = true
  try {
    const response = await api.query({
      query: query.value,
      document_ids: selectedDocIds.value
    })

    answer.value = response.data.answer
    citations.value = response.data.citations
  } catch (error) {
    console.error('Query error:', error)
    const errorMsg = error.response?.data?.error || error.message || 'Query failed'
    MessagePlugin.error('Query failed: ' + errorMsg)
  } finally {
    loading.value = false
  }
}

const handleViewSource = async ({ documentId, pageNumber, bbox }) => {
  try {
    console.log('View source called:', { documentId, pageNumber, bbox })

    // Get document info for total pages
    const doc = documents.value.find(d => d.id === documentId)

    viewerConfig.value = {
      documentId,
      pageNumber,
      bbox,
      totalPages: doc?.total_pages || 1
    }
    console.log('Viewer config set:', viewerConfig.value)
    MessagePlugin.success('Loading page ' + pageNumber)
  } catch (error) {
    console.error('Failed to load PDF:', error)
    MessagePlugin.error('Failed to load PDF: ' + error.message)
  }
}

const handleUpload = async () => {
  if (!uploadFiles.value || uploadFiles.value.length === 0) {
    MessagePlugin.warning('Please select a file')
    return
  }

  try {
    loading.value = true
    const file = uploadFiles.value[0]
    const actualFile = file.raw || file

    console.log('=== UPLOAD START ===')
    console.log('File name:', actualFile.name)
    console.log('File size:', actualFile.size, 'bytes', `(${(actualFile.size / (1024 * 1024)).toFixed(2)} MB)`)
    console.log('File type:', actualFile.type)

    const startTime = Date.now()
    await api.uploadDocument(actualFile)
    const duration = Date.now() - startTime

    console.log('=== UPLOAD SUCCESS ===')
    console.log('Upload took:', duration, 'ms')

    MessagePlugin.success('PDF uploaded successfully')
    showUploadDialog.value = false
    uploadFiles.value = []
    await loadDocuments()
  } catch (error) {
    console.error('=== UPLOAD ERROR ===')
    console.error('Error object:', error)
    console.error('Response status:', error.response?.status)
    console.error('Response data:', error.response?.data)
    console.error('Error message:', error.message)

    MessagePlugin.error('Upload failed: ' + (error.response?.data?.error || error.response?.data?.message || error.message))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadDocuments()
})
</script>

<style scoped>
.main-container {
  height: 100vh;
}

.header {
  background: #0052d9;
  color: white;
  padding: 16px 24px;
}

.sidebar {
  background: #f5f5f5;
  padding: 16px;
}

.document-list {
  margin-top: 16px;
  max-height: calc(100vh - 200px);
  overflow-y: auto;
}

.main-content {
  padding: 16px;
}

.chat-panel,
.viewer-panel {
  height: calc(100vh - 100px);
  overflow-y: auto;
}

.answer-section,
.citations-section {
  margin-top: 20px;
}
</style>
