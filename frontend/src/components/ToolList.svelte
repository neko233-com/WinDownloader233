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
      <div class="loading-ring">
        <div class="ring-segment"></div>
      </div>
      <p class="empty-label">{ui['loading'] || 'Loading'}</p>
    </div>
  {:else if tools.length === 0}
    <div class="empty-state">
      <div class="empty-glyph">
        <svg viewBox="0 0 24 24" width="40" height="40" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8"/>
          <path d="m21 21-4.35-4.35"/>
          <path d="M8 11h6"/>
        </svg>
      </div>
      <p class="empty-label">{ui['no_results'] || 'No results'}</p>
    </div>
  {:else}
    <div class="tool-grid">
      {#each tools as tool, i (tool.id)}
        <div class="card-enter" style="animation-delay: {i * 30}ms">
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
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .tool-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px 28px 28px;
    scroll-behavior: smooth;
  }

  .tool-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
    gap: 14px;
  }

  .card-enter {
    animation: cardIn 0.4s var(--ease-spring) both;
  }

  @keyframes cardIn {
    from {
      opacity: 0;
      transform: translateY(12px) scale(0.97);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 60%;
    gap: 16px;
  }

  .empty-glyph {
    color: var(--text-tertiary);
    opacity: 0.4;
  }

  .empty-label {
    font-size: 14px;
    color: var(--text-tertiary);
    font-weight: 500;
  }

  .loading-ring {
    width: 36px;
    height: 36px;
    position: relative;
  }

  .ring-segment {
    position: absolute;
    inset: 0;
    border: 2.5px solid transparent;
    border-top-color: var(--accent-blue);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  /* Scrollbar */
  .tool-list::-webkit-scrollbar { width: 5px; }
  .tool-list::-webkit-scrollbar-track { background: transparent; }
  .tool-list::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.08);
    border-radius: var(--radius-full);
  }
  .tool-list::-webkit-scrollbar-thumb:hover {
    background: rgba(255, 255, 255, 0.15);
  }
</style>
