import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

const apiClient = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  headers: {
    'Content-Type': 'application/json'
  }
})

export default {
  // Documents
  async getDocuments() {
    return apiClient.get('/documents')
  },

  async uploadDocument(file: File) {
    console.log('[API] Uploading document:', file.name, 'size:', file.size)
    const formData = new FormData()
    formData.append('file', file)

    console.log('[API] FormData created, sending POST request...')
    try {
      const response = await apiClient.post('/documents/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
        onUploadProgress: (progressEvent) => {
          const percentCompleted = Math.round((progressEvent.loaded * 100) / (progressEvent.total || 1))
          console.log(`[API] Upload progress: ${percentCompleted}% (${progressEvent.loaded}/${progressEvent.total} bytes)`)
        }
      })
      console.log('[API] Upload response:', response.status, response.data)
      return response
    } catch (error) {
      console.error('[API] Upload error:', error)
      throw error
    }
  },

  async getDocument(id: string) {
    return apiClient.get(`/documents/${id}`)
  },

  async deleteDocument(id: string) {
    return apiClient.delete(`/documents/${id}`)
  },

  async getDocumentFileUrl(id: string): Promise<string> {
    // Return URL to PDF file
    return `${API_BASE_URL}/api/v1/documents/${id}/file`
  },

  // Chat
  async query(data: { query: string; document_ids: string[] }) {
    return apiClient.post('/chat/query', data)
  }
}
