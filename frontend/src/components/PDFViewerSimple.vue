<template>
  <div class="pdf-viewer-simple">
    <div class="controls">
      <t-space>
        <t-tag theme="primary">{{ filename || 'PDF Document' }}</t-tag>
        <t-tag v-if="pageNumber">Page {{ pageNumber }}</t-tag>
        <t-button size="small" theme="default" @click="openInNewTab">
          Open in New Tab
        </t-button>
      </t-space>
    </div>
    <div class="viewer-container">
      <iframe
        ref="pdfFrame"
        :src="pdfUrl + '#page=' + (pageNumber || 1)"
        class="pdf-iframe"
        @load="onLoad"
      />
      <div v-if="loading" class="loading-overlay">
        <t-loading size="large" text="Loading PDF..." />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { MessagePlugin } from 'tdesign-vue-next'

const props = defineProps<{
  pdfUrl: string
  pageNumber?: number
  bbox?: { x1: number; y1: number; x2: number; y2: number } | null
  filename?: string
}>()

const pdfFrame = ref<HTMLIFrameElement>()
const loading = ref(true)

const onLoad = () => {
  loading.value = false
  console.log('PDF loaded successfully')
}

const openInNewTab = () => {
  const url = props.pdfUrl + '#page=' + (props.pageNumber || 1)
  window.open(url, '_blank')
}

// Watch for URL changes
watch(() => props.pdfUrl, () => {
  loading.value = true
  console.log('Loading new PDF:', props.pdfUrl)
})

// Watch for page changes
watch(() => props.pageNumber, (newPage) => {
  if (newPage && pdfFrame.value) {
    pdfFrame.value.src = props.pdfUrl + '#page=' + newPage
    console.log('Navigating to page:', newPage)
  }
})
</script>

<style scoped>
.pdf-viewer-simple {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #f5f5f5;
}

.controls {
  padding: 12px;
  background: white;
  border-bottom: 1px solid #e0e0e0;
}

.viewer-container {
  flex: 1;
  position: relative;
  overflow: hidden;
}

.pdf-iframe {
  width: 100%;
  height: 100%;
  border: none;
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
