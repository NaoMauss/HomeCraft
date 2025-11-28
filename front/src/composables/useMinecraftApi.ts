import { ref } from 'vue'
import { ofetch } from 'ofetch'

// Use relative URL so it works from any domain (local, public IP, or Playit.gg)
// This will make requests to the same host/domain that served the frontend
const API_BASE_URL = import.meta.env.VITE_API_URL || ''

export interface MinecraftServer {
  name: string
  namespace: string
  eula: boolean
  memory: string
  storageSize: string
  version: string
  serverType: string
  maxPlayers: number
  difficulty: string
  gamemode: string
  phase?: string
  endpoint?: string
  sftpEndpoint?: string
  sftpUsername?: string
  sftpPassword?: string
  allocatedMemory?: string
  createdAt?: string
}

export interface CreateServerRequest {
  name: string
  eula: boolean
  memory: string
  storageSize: string
  version?: string
  serverType?: string
  maxPlayers?: number
  difficulty?: string
  gamemode?: string
}

export function useMinecraftApi() {
  const loading = ref(false)
  const error = ref<string | null>(null)

  const createServer = async (request: CreateServerRequest): Promise<MinecraftServer> => {
    loading.value = true
    error.value = null
    try {
      const response = await ofetch<MinecraftServer>(`${API_BASE_URL}/api/v1/servers`, {
        method: 'POST',
        body: request,
      })
      return response
    } catch (e: any) {
      error.value = e.data?.message || e.message || 'Failed to create server'
      throw e
    } finally {
      loading.value = false
    }
  }

  const listServers = async (): Promise<MinecraftServer[]> => {
    loading.value = true
    error.value = null
    try {
      const response = await ofetch<{ count: number; items: MinecraftServer[] }>(`${API_BASE_URL}/api/v1/servers`)
      return response.items
    } catch (e: any) {
      error.value = e.data?.message || e.message || 'Failed to fetch servers'
      throw e
    } finally {
      loading.value = false
    }
  }

  const getServer = async (name: string): Promise<MinecraftServer> => {
    loading.value = true
    error.value = null
    try {
      const response = await ofetch<MinecraftServer>(`${API_BASE_URL}/api/v1/servers/${name}`)
      return response
    } catch (e: any) {
      error.value = e.data?.message || e.message || 'Failed to fetch server'
      throw e
    } finally {
      loading.value = false
    }
  }

  const deleteServer = async (name: string): Promise<void> => {
    loading.value = true
    error.value = null
    try {
      await ofetch(`${API_BASE_URL}/api/v1/servers/${name}`, {
        method: 'DELETE',
      })
    } catch (e: any) {
      error.value = e.data?.message || e.message || 'Failed to delete server'
      throw e
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    createServer,
    listServers,
    getServer,
    deleteServer,
  }
}
