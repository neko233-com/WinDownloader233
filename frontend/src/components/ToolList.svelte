<script lang="ts">
  import type { Tool, Progress } from '../lib/types';
  import ToolCard from './ToolCard.svelte';

  interface Props {
    tools: Tool[];
    progressMap: Map<string, Progress>;
    lang: string;
    loading: boolean;
    ui: Record<string, string>;
    getToolName: (tool: Tool) => string;
    getToolDesc: (tool: Tool) => string;
    onInstall: (tool: Tool) => void;
    onUninstall: (tool: Tool) => void;
  }

  let { tools, progressMap, lang, loading, ui, getToolName, getToolDesc, onInstall, onUninstall }: Props = $props();
</script>

<div class="tool-list">
  {#if loading && tools.length === 0}
    <div class="empty-state">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>
  {:else if tools.length === 0}
    <div class="empty-state">
      <span class="empty-icon">🔍</span>
      <p>{ui['no_results'] || '没有找到匹配的工具'}</p>
    </div>
  {:else}
    <div class="tool-grid">
      {#each tools as tool (tool.id)}
        <ToolCard
          {tool}
          progress={progressMap.get(tool.id)}
          {lang}
          {ui}
          name={getToolName(tool)}
          description={getToolDesc(tool)}
          onInstall={() => onInstall(tool)}
          onUninstall={() => onUninstall(tool)}
        />
      {/each}
    </div>
  {/if}
</div>

<style>
  .tool-list {
    flex: 1;
    overflow-y: auto;
    padding: 20px 24px;
  }

  .tool-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 12px;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 60%;
    color: #6e7681;
    gap: 12px;
  }

  .empty-icon {
    font-size: 48px;
    opacity: 0.5;
  }

  .loading-spinner {
    width: 32px;
    height: 32px;
    border: 3px solid #2d3148;
    border-top-color: #58a6ff;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  /* Custom scrollbar */
  .tool-list::-webkit-scrollbar {
    width: 6px;
  }
  .tool-list::-webkit-scrollbar-track {
    background: transparent;
  }
  .tool-list::-webkit-scrollbar-thumb {
    background: #2d3148;
    border-radius: 3px;
  }
  .tool-list::-webkit-scrollbar-thumb:hover {
    background: #484f58;
  }
</style>
