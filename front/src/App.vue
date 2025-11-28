<script setup lang="ts">
import { LucidePickaxe, LucideSword, LucideMap } from 'lucide-vue-next';
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
          <button class="flex items-center gap-2 border-3 border-black bg-mc-diamond px-6 py-2 font-bold text-black shadow-neo-sm transition-all hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-none active:bg-mc-diamond/80">
            <LucideMap class="h-5 w-5" />
            Servers
          </button>
          <button class="flex items-center gap-2 border-3 border-black bg-mc-gold px-6 py-2 font-bold text-black shadow-neo-sm transition-all hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-none active:bg-mc-gold/80">
            <LucideSword class="h-5 w-5" />
            Login
          </button>
        </nav>
      </div>
    </header>

    <!-- Main Content -->
    <main class="mx-auto max-w-6xl">
      <div class="grid gap-8 md:grid-cols-2 lg:grid-cols-3">
        <!-- Welcome Card -->
        <div class="col-span-full border-3 border-black bg-white p-8 shadow-neo-lg">
          <h2 class="mb-4 text-3xl font-bold text-black">Welcome back, Miner!</h2>
          <p class="mb-6 text-lg text-gray-700">Manage your Minecraft servers with style. Deploy, monitor, and configure your worlds with our blocky interface.</p>
          <div class="flex gap-4">
            <button class="border-3 border-black bg-mc-redstone px-8 py-3 text-xl font-bold text-white shadow-neo transition-all hover:translate-x-[4px] hover:translate-y-[4px] hover:shadow-none">
              Create Server
            </button>
          </div>
        </div>

        <!-- Stats Card 1 -->
        <div class="border-3 border-black bg-mc-stone p-6 shadow-neo text-white">
          <h3 class="mb-2 text-xl font-bold">Active Servers</h3>
          <p class="text-4xl font-mono">12</p>
        </div>

        <!-- Stats Card 2 -->
        <div class="border-3 border-black bg-mc-dirt p-6 shadow-neo text-white">
          <h3 class="mb-2 text-xl font-bold">Players Online</h3>
          <p class="text-4xl font-mono">1,337</p>
        </div>

        <!-- Stats Card 3 -->
        <div class="border-3 border-black bg-mc-obsidian p-6 shadow-neo text-white">
          <h3 class="mb-2 text-xl font-bold">Uptime</h3>
          <p class="text-4xl font-mono">99.9%</p>
        </div>
      </div>
    </main>
  </div>
</template>

<style>
/* Global overrides if needed */
</style>
