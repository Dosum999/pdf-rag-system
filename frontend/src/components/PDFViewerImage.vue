<template>
  <div class="pdf-viewer-image">
    <div class="controls">
      <t-button size="small" @click="previousPage" :disabled="currentPage <= 1">이전</t-button>
      <span>{{ currentPage }} / {{ totalPages }} 쪽</span>
      <t-button size="small" @click="nextPage" :disabled="currentPage >= totalPages">다음</t-button>
      <t-button size="small" @click="zoomIn">+</t-button>
      <t-button size="small" @click="zoomOut">-</t-button>
    </div>
    <div class="image-wrapper" ref="imageWrapper">
      <img
        v-if="imageUrl"
        :src="imageUrl"
        :style="{ width: `${zoom * 100}%` }"
        @load="onImageLoad"
        @error="onImageError"
      />
      <div v-if="loading" class="loading-overlay">
        <t-loading size="large" text="페이지 로딩 중..." />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { MessagePlugin } from 'tdesign-vue-next'

const props = defineProps<{
  documentId: string
  pageNumber?: number
  bbox?: { x1: number; y1: number; x2: number; y2: number } | null
  totalPages?: number
}>()

const currentPage = ref(props.pageNumber || 1)
const zoom = ref(1)
const loading = ref(false)
const imageWrapper = ref<HTMLDivElement>()

const imageUrl = computed(() => {
  if (!props.documentId) return null

  const baseUrl = `http://localhost:8080/api/v1/documents/${props.documentId}/page/${currentPage.value}/image`

  if (props.bbox) {
    const params = new URLSearchParams({
      bbox_x1: props.bbox.x1.toString(),
      bbox_y1: props.bbox.y1.toString(),
      bbox_x2: props.bbox.x2.toString(),
      bbox_y2: props.bbox.y2.toString()
    })
    return `${baseUrl}?${params.toString()}`
  }

  return baseUrl
})

const totalPages = computed(() => props.totalPages || 1)

const onImageLoad = () => {
  loading.value = false
  console.log('Page image loaded:', currentPage.value)
}

const onImageError = () => {
  loading.value = false
  console.error('Failed to load page image:', currentPage.value)
  MessagePlugin.error('Failed to load page image')
}

const previousPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
    loading.value = true
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    loading.value = true
  }
}

const zoomIn = () => {
  zoom.value = Math.min(3, zoom.value + 0.25)
}

const zoomOut = () => {
  zoom.value = Math.max(0.5, zoom.value - 0.25)
}

// Watch for page number changes from props
watch(() => props.pageNumber, (newPage) => {
  if (newPage && newPage !== currentPage.value) {
    currentPage.value = newPage
    loading.value = true
  }
})

// Watch for document ID changes
watch(() => props.documentId, () => {
  currentPage.value = props.pageNumber || 1
  zoom.value = 1
  loading.value = true
})
</script>

<style scoped>
.pdf-viewer-image {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #f5f5f5;
}

.controls {
  padding: 12px;
  display: flex;
  gap: 8px;
  align-items: center;
  border-bottom: 1px solid #e0e0e0;
  background: white;
}

.image-wrapper {
  flex: 1;
  overflow: auto;
  position: relative;
  padding: 20px;
  display: flex;
  justify-content: center;
  align-items: flex-start;
}

.image-wrapper img {
  max-width: 100%;
  height: auto;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  background: white;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
