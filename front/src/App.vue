<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { LucidePickaxe, LucideServer, LucideRefreshCw, LucidePlus, LucideX, LucideTrash2, LucideCopy } from 'lucide-vue-next'
import { useMinecraftApi, type MinecraftServer, type CreateServerRequest } from './composables/useMinecraftApi'

const { createServer, listServers, deleteServer, loading, error } = useMinecraftApi()

const servers = ref<MinecraftServer[]>([])
const showCreateModal = ref(false)
const refreshing = ref(false)
const copySuccess = ref(false)

const newServer = ref<CreateServerRequest>({
  name: '',
  eula: false,
  memory: '2Gi',
  storageSize: '5Gi',
  version: 'LATEST',
  serverType: 'VANILLA',
  maxPlayers: 20,
  difficulty: 'normal',
  gamemode: 'survival',
})

const loadServers = async () => {
  refreshing.value = true
  try {
    servers.value = await listServers()
  } catch (e) {
    console.error('Failed to load servers:', e)
  } finally {
    refreshing.value = false
  }
}

const handleCreateServer = async () => {
  if (!newServer.value.eula) {
    alert('You must accept the Minecraft EULA to create a server')
    return
  }

  try {
    const created = await createServer(newServer.value)
    servers.value.push(created)
    showCreateModal.value = false
    resetForm()
    await loadServers()
  } catch (e: any) {
    alert(error.value || 'Failed to create server')
  }
}

const handleDeleteServer = async (name: string) => {
  if (!confirm(`Are you sure you want to delete server "${name}"?`)) {
    return
  }

  try {
    await deleteServer(name)
    servers.value = servers.value.filter(s => s.name !== name)
  } catch (e) {
    alert(error.value || 'Failed to delete server')
  }
}

const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    copySuccess.value = true
    setTimeout(() => {copySuccess.value = false}, 2000)
  } catch (e) {
    alert('Failed to copy to clipboard')
  }
}

const resetForm = () => {
  newServer.value = {
    name: '',
    eula: false,
    memory: '2Gi',
    storageSize: '5Gi',
    version: 'LATEST',
    serverType: 'VANILLA',
    maxPlayers: 20,
    difficulty: 'normal',
    gamemode: 'survival',
  }
}

const getPhaseColor = (phase?: string) => {
  switch (phase) {
    case 'Running': return 'bg-green-500'
    case 'Starting': return 'bg-yellow-500'
    case 'Pending': return 'bg-gray-500'
    case 'Failed': return 'bg-red-500'
    default: return 'bg-gray-400'
  }
}

onMounted(() => {
  loadServers()
  // Refresh servers every 10 seconds
  setInterval(loadServers, 10000)
})
</script>

<template>
  <div class="min-h-screen bg-bg p-8 font-sans">
    <!-- Header -->
    <header class="mb-12 border-3 border-black bg-mc-grass p-6 shadow-neo">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-4">
          <div class="flex h-12 w-12 items-center justify-center border-3 border-black bg-white shadow-neo-sm">
            <LucidePickaxe class="h-8 w-8 text-black" />
          </div>
          <h1 class="text-4xl font-bold uppercase tracking-wider text-black">HomeCraft</h1>
        </div>
        <nav class="flex gap-4">
          <button 
            @click="loadServers"
            :disabled="refreshing"
            class="flex items-center gap-2 border-3 border-black bg-mc-diamond px-6 py-2 font-bold text-black shadow-neo-sm transition-all hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-none active:bg-mc-diamond/80 disabled:opacity-50">
            <LucideRefreshCw class="h-5 w-5" :class="{'animate-spin': refreshing}" />
            Refresh
          </button>
          <button 
            @click="showCreateModal = true"
            class="flex items-center gap-2 border-3 border-black bg-mc-redstone px-6 py-2 font-bold text-white shadow-neo-sm transition-all hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-none active:bg-mc-redstone/80">
            <LucidePlus class="h-5 w-5" />
            Create Server
          </button>
        </nav>
      </div>
    </header>

    <!-- Main Content -->
    <main class="mx-auto max-w-7xl">
      <!-- Server Stats -->
      <div class="grid gap-8 md:grid-cols-3 mb-8">
        <div class="border-3 border-black bg-mc-stone p-6 shadow-neo text-white">
          <h3 class="mb-2 text-xl font-bold">Total Servers</h3>
          <p class="text-4xl font-mono">{{ servers.length }}</p>
        </div>

        <div class="border-3 border-black bg-mc-dirt p-6 shadow-neo text-white">
          <h3 class="mb-2 text-xl font-bold">Running Servers</h3>
          <p class="text-4xl font-mono">{{ servers.filter(s => s.phase === 'Running').length }}</p>
        </div>

        <div class="border-3 border-black bg-mc-obsidian p-6 shadow-neo text-white">
          <h3 class="mb-2 text-xl font-bold">Total Players</h3>
          <p class="text-4xl font-mono">{{ servers.reduce((sum, s) => sum + (s.maxPlayers || 0), 0) }}</p>
        </div>
      </div>

      <!-- Servers List -->
      <div class="border-3 border-black bg-white p-8 shadow-neo-lg">
        <h2 class="mb-6 text-3xl font-bold text-black flex items-center gap-3">
          <LucideServer class="h-8 w-8" />
          Your Servers
        </h2>

        <div v-if="servers.length === 0" class="text-center py-12 text-gray-500">
          <p class="text-xl mb-4">No servers yet!</p>
          <p>Click "Create Server" to get started</p>
        </div>

        <div v-else class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          <div 
            v-for="server in servers" 
            :key="server.name"
            class="border-3 border-black bg-gray-50 p-6 shadow-neo-sm hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-none transition-all">
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-xl font-bold text-black">{{ server.name }}</h3>
              <span 
                :class="getPhaseColor(server.phase)"
                class="px-3 py-1 text-xs font-bold text-white border-2 border-black">
                {{ server.phase || 'Unknown' }}
              </span>
            </div>

            <div class="space-y-2 text-sm mb-4">
              <p><span class="font-bold">Version:</span> {{ server.version }}</p>
              <p><span class="font-bold">Type:</span> {{ server.serverType }}</p>
              <p><span class="font-bold">Memory:</span> {{ server.memory }}</p>
              <p><span class="font-bold">Max Players:</span> {{ server.maxPlayers }}</p>
              
              <div v-if="server.endpoint" class="pt-2 border-t-2 border-gray-300">
                <p class="font-bold text-green-600 mb-1">Connection Info:</p>
                <div class="flex items-center gap-2">
                  <code class="bg-black text-green-400 px-2 py-1 text-xs flex-1 font-mono">{{ server.endpoint }}</code>
                  <button 
                    @click="copyToClipboard(server.endpoint!)"
                    class="border-2 border-black bg-mc-gold p-1 hover:bg-mc-gold/80 transition-colors">
                    <LucideCopy class="h-4 w-4" />
                  </button>
                </div>
              </div>
            </div>

            <button 
              @click="handleDeleteServer(server.name)"
              :disabled="loading"
              class="w-full flex items-center justify-center gap-2 border-3 border-black bg-red-500 px-4 py-2 font-bold text-white shadow-neo-sm transition-all hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-none active:bg-red-600 disabled:opacity-50">
              <LucideTrash2 class="h-4 w-4" />
              Delete
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- Create Server Modal -->
    <div 
      v-if="showCreateModal" 
      class="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50"
      @click.self="showCreateModal = false">
      <div class="border-3 border-black bg-white p-8 shadow-neo-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between mb-6">
          <h2 class="text-3xl font-bold text-black">Create Minecraft Server</h2>
          <button 
            @click="showCreateModal = false"
            class="border-3 border-black bg-red-500 p-2 text-white hover:bg-red-600 transition-colors">
            <LucideX class="h-6 w-6" />
          </button>
        </div>

        <form @submit.prevent="handleCreateServer" class="space-y-4">
          <div>
            <label class="block font-bold mb-2">Server Name *</label>
            <input 
              v-model="newServer.name"
              required
              pattern="[a-z0-9]([-a-z0-9]*[a-z0-9])?"
              placeholder="my-server"
              class="w-full border-3 border-black p-3 font-mono focus:outline-none focus:ring-4 focus:ring-mc-gold"
            />
            <p class="text-xs text-gray-600 mt-1">Lowercase letters, numbers, and hyphens only</p>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block font-bold mb-2">Memory *</label>
              <select 
                v-model="newServer.memory"
                required
                class="w-full border-3 border-black p-3 font-mono focus:outline-none focus:ring-4 focus:ring-mc-gold">
                <option value="1Gi">1 GB</option>
                <option value="2Gi">2 GB</option>
                <option value="4Gi">4 GB</option>
                <option value="8Gi">8 GB</option>
              </select>
            </div>

            <div>
              <label class="block font-bold mb-2">Storage Size *</label>
              <select 
                v-model="newServer.storageSize"
                required
                class="w-full border-3 border-black p-3 font-mono focus:outline-none focus:ring-4 focus:ring-mc-gold">
                <option value="1Gi">1 GB</option>
                <option value="5Gi">5 GB</option>
                <option value="10Gi">10 GB</option>
                <option value="20Gi">20 GB</option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block font-bold mb-2">Version</label>
              <input 
                v-model="newServer.version"
                placeholder="LATEST"
                class="w-full border-3 border-black p-3 font-mono focus:outline-none focus:ring-4 focus:ring-mc-gold"
              />
            </div>

            <div>
              <label class="block font-bold mb-2">Server Type</label>
              <select 
                v-model="newServer.serverType"
                class="w-full border-3 border-black p-3 font-mono focus:outline-none focus:ring-4 focus:ring-mc-gold">
                <option value="VANILLA">Vanilla</option>
                <option value="PAPER">Paper</option>
                <option value="FORGE">Forge</option>
                <option value="FABRIC">Fabric</option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-3 gap-4">
            <div>
              <label class="block font-bold mb-2">Max Players</label>
              <input 
                v-model.number="newServer.maxPlayers"
                type="number"
                min="1"
                max="1000"
                class="w-full border-3 border-black p-3 font-mono focus:outline-none focus:ring-4 focus:ring-mc-gold"
              />
            </div>

            <div>
              <label class="block font-bold mb-2">Difficulty</label>
              <select 
                v-model="newServer.difficulty"
                class="w-full border-3 border-black p-3 font-mono focus:outline-none focus:ring-4 focus:ring-mc-gold">
                <option value="peaceful">Peaceful</option>
                <option value="easy">Easy</option>
                <option value="normal">Normal</option>
                <option value="hard">Hard</option>
              </select>
            </div>

            <div>
              <label class="block font-bold mb-2">Gamemode</label>
              <select 
                v-model="newServer.gamemode"
                class="w-full border-3 border-black p-3 font-mono focus:outline-none focus:ring-4 focus:ring-mc-gold">
                <option value="survival">Survival</option>
                <option value="creative">Creative</option>
                <option value="adventure">Adventure</option>
                <option value="spectator">Spectator</option>
              </select>
            </div>
          </div>

          <div class="border-3 border-black bg-yellow-100 p-4">
            <label class="flex items-start gap-3 cursor-pointer">
              <input 
                v-model="newServer.eula"
                type="checkbox"
                required
                class="mt-1 h-5 w-5 border-3 border-black"
              />
              <span class="text-sm">
                <span class="font-bold">I accept the Minecraft EULA *</span><br>
                By checking this box, you agree to the 
                <a href="https://account.mojang.com/documents/minecraft_eula" target="_blank" class="text-blue-600 underline">
                  Minecraft End User License Agreement
                </a>
              </span>
            </label>
          </div>

          <div class="flex gap-4 pt-4">
            <button 
              type="submit"
              :disabled="loading || !newServer.eula"
              class="flex-1 border-3 border-black bg-mc-redstone px-6 py-3 text-xl font-bold text-white shadow-neo transition-all hover:translate-x-[4px] hover:translate-y-[4px] hover:shadow-none disabled:opacity-50 disabled:cursor-not-allowed">
              {{ loading ? 'Creating...' : 'Create Server' }}
            </button>
            <button 
              type="button"
              @click="showCreateModal = false"
              :disabled="loading"
              class="border-3 border-black bg-gray-300 px-6 py-3 text-xl font-bold text-black shadow-neo transition-all hover:translate-x-[4px] hover:translate-y-[4px] hover:shadow-none disabled:opacity-50">
              Cancel
            </button>
          </div>

          <p v-if="error" class="text-red-600 font-bold text-center">{{ error }}</p>
        </form>
      </div>
    </div>

    <!-- Copy Success Toast -->
    <div 
      v-if="copySuccess"
      class="fixed bottom-8 right-8 border-3 border-black bg-green-500 text-white px-6 py-3 font-bold shadow-neo-lg z-50">
      Copied to clipboard!
    </div>
  </div>
</template>
