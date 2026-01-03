<template>
  <div class="pdf-viewer">
    <div class="controls">
      <t-button size="small" @click="previousPage" :disabled="currentPage <= 1">Previous</t-button>
      <span>Page {{ currentPage }} / {{ totalPages }}</span>
      <t-button size="small" @click="nextPage" :disabled="currentPage >= totalPages">Next</t-button>
      <t-button size="small" @click="zoomIn">+</t-button>
      <t-button size="small" @click="zoomOut">-</t-button>
    </div>
    <div class="canvas-wrapper" ref="canvasWrapper">
      <canvas ref="canvas"></canvas>
      <div v-if="highlightBox" class="highlight" :style="highlightStyle"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed, nextTick } from 'vue'
import { MessagePlugin } from 'tdesign-vue-next'
import * as pdfjsLib from 'pdfjs-dist'

// Configure pdf.js worker - match the installed version
pdfjsLib.GlobalWorkerOptions.workerSrc = `https://cdnjs.cloudflare.com/ajax/libs/pdf.js/${pdfjsLib.version}/pdf.worker.min.js`

const props = defineProps<{
  pdfUrl: string
  pageNumber?: number
  bbox?: { x1: number; y1: number; x2: number; y2: number } | null
}>()

const canvas = ref<HTMLCanvasElement>()
const currentPage = ref(1)
const totalPages = ref(0)
const scale = ref(1.5)
const pdfDoc = ref<any>(null)
const loading = ref(false)

const highlightBox = computed(() => props.bbox)

const highlightStyle = computed(() => {
  if (!highlightBox.value) return {}
  const bbox = highlightBox.value
  return {
    left: `${bbox.x1 * scale.value}px`,
    top: `${bbox.y1 * scale.value}px`,
    width: `${(bbox.x2 - bbox.x1) * scale.value}px`,
    height: `${(bbox.y2 - bbox.y1) * scale.value}px`
  }
})

const loadPDF = async () => {
  try {
    loading.value = true
    console.log('Loading PDF from:', props.pdfUrl)

    // Clean up previous PDF to prevent memory leak
    if (pdfDoc.value) {
      await pdfDoc.value.destroy()
      pdfDoc.value = null
    }

    const loadingTask = pdfjsLib.getDocument({
      url: props.pdfUrl,
      withCredentials: false,
      isEvalSupported: false
    })

    pdfDoc.value = await loadingTask.promise
    console.log('PDF loaded successfully, pages:', pdfDoc.value.numPages)
    totalPages.value = pdfDoc.value.numPages

    if (props.pageNumber && props.pageNumber > 0 && props.pageNumber <= totalPages.value) {
      currentPage.value = props.pageNumber
    } else {
      currentPage.value = 1
    }

    await renderPage(currentPage.value)
  } catch (error) {
    console.error('Failed to load PDF:', error)
    MessagePlugin.error('Failed to load PDF file: ' + error.message)
  } finally {
    loading.value = false
  }
}

const renderPage = async (pageNum: number) => {
  if (!pdfDoc.value) {
    console.error('No PDF document loaded')
    return
  }

  if (!canvas.value) {
    console.error('Canvas element not ready')
    return
  }

  try {
    console.log('Rendering page:', pageNum, 'of', totalPages.value)
    const page = await pdfDoc.value.getPage(pageNum)
    console.log('Page loaded:', pageNum)

    const viewport = page.getViewport({ scale: scale.value })
    const context = canvas.value.getContext('2d')
    if (!context) {
      console.error('Failed to get canvas context')
      return
    }

    canvas.value.height = viewport.height
    canvas.value.width = viewport.width

    const renderContext = {
      canvasContext: context,
      viewport: viewport
    }

    await page.render(renderContext).promise
    console.log('Page rendered successfully:', pageNum)
  } catch (error: any) {
    // Handle "Cannot read from private field" gracefully
    console.error('Failed to render page:', pageNum, error)

    // Try to render page 1 if current page fails
    if (pageNum !== 1) {
      console.log('Attempting to render page 1 instead...')
      try {
        currentPage.value = 1
        await renderPage(1)
        MessagePlugin.warning(`Cannot render page ${pageNum}. Showing page 1 instead.`)
      } catch (fallbackError) {
        MessagePlugin.error('Cannot render PDF. The file may be corrupted.')
      }
    } else {
      MessagePlugin.error('Cannot render PDF. The file may be corrupted.')
    }
  }
}

const previousPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
    renderPage(currentPage.value)
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    renderPage(currentPage.value)
  }
}

const zoomIn = () => {
  scale.value += 0.25
  renderPage(currentPage.value)
}

const zoomOut = () => {
  scale.value = Math.max(0.5, scale.value - 0.25)
  renderPage(currentPage.value)
}

// Fix: Watch pdfUrl to reload when document changes
watch(() => props.pdfUrl, () => {
  loadPDF()
})

// Fix: Watch pageNumber to update when citation changes
watch(() => props.pageNumber, (newPage) => {
  if (newPage && pdfDoc.value) {
    currentPage.value = newPage
    renderPage(currentPage.value)
  }
})

onMounted(async () => {
  await nextTick()
  if (props.pdfUrl) {
    loadPDF()
  }
})

// Fix: Clean up PDF document on unmount to prevent memory leak
onUnmounted(() => {
  if (pdfDoc.value) {
    pdfDoc.value.destroy()
  }
})
</script>

<style scoped>
.pdf-viewer {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.controls {
  padding: 12px;
  display: flex;
  gap: 8px;
  align-items: center;
  border-bottom: 1px solid #e0e0e0;
}

.canvas-wrapper {
  flex: 1;
  overflow: auto;
  position: relative;
  background: #f5f5f5;
  padding: 20px;
}

.highlight {
  position: absolute;
  border: 2px solid #0052d9;
  background: rgba(0, 82, 217, 0.1);
  pointer-events: none;
}
</style>
