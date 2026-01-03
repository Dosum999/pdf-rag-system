<template>
  <t-card :bordered="true" class="citation-card" hover-shadow>
    <template #header>
      <div class="header">
        <FileIcon />
        <div>
          <div class="filename">{{ citation.filename }}</div>
          <t-tag size="small" theme="primary">{{ citation.page_number }}쪽</t-tag>
          <t-tag v-if="hasBbox" size="small" theme="success">위치 정보 있음</t-tag>
        </div>
      </div>
    </template>

    <div class="content">{{ citation.content }}</div>

    <div class="actions">
      <t-button size="small" theme="primary" @click="handleView">
        원문 보기
      </t-button>
      <t-tag size="small">유사도: {{ citation.score.toFixed(2) }}</t-tag>
    </div>
  </t-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { FileIcon } from 'tdesign-icons-vue-next'

const props = defineProps<{
  citation: {
    document_id: string
    filename: string
    page_number: number
    content: string
    bbox_x1?: number
    bbox_y1?: number
    bbox_x2?: number
    bbox_y2?: number
    score: number
  }
}>()

const emit = defineEmits<{
  (e: 'viewSource', data: { documentId: string; pageNumber: number; bbox: { x1: number; y1: number; x2: number; y2: number } | null }): void
}>()

const hasBbox = computed(() =>
  props.citation.bbox_x1 !== undefined &&
  props.citation.bbox_y1 !== undefined &&
  props.citation.bbox_x2 !== undefined &&
  props.citation.bbox_y2 !== undefined
)

const handleView = () => {
  const bbox = hasBbox.value ? {
    x1: props.citation.bbox_x1!,
    y1: props.citation.bbox_y1!,
    x2: props.citation.bbox_x2!,
    y2: props.citation.bbox_y2!
  } : null

  emit('viewSource', {
    documentId: props.citation.document_id,
    pageNumber: props.citation.page_number,
    bbox
  })
}
</script>

<style scoped>
.citation-card {
  margin-bottom: 12px;
}

.header {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.filename {
  font-weight: 500;
  margin-bottom: 4px;
}

.content {
  margin: 12px 0;
  line-height: 1.6;
  max-height: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.actions {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #e0e0e0;
}
</style>
