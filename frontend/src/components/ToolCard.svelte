<script lang="ts">
  import type { Tool, Progress } from '../lib/types';

  interface Props {
    tool: Tool;
    progress?: Progress;
    lang: string;
    ui: Record<string, string>;
    name: string;
    description: string;
    onInstall: () => void;
    onUninstall: () => void;
  }

  let { tool, progress, lang, ui, name, description, onInstall, onUninstall }: Props = $props();

  let installing = $derived(progress?.status === 'installing' || progress?.status === 'downloading');
  let done = $derived(progress?.status === 'done');
  let error = $derived(progress?.status === 'error');

  function openHomepage() {
    if (tool.homepage) {
      window.open(tool.homepage, '_blank');
    }
  }
</script>

<div class="tool-card" class:installing>
  <div class="card-header">
    <div class="tool-icon">
      {#if tool.icon === 'vscode'}
        <span style="color:#007acc">⚡</span>
      {:else if tool.icon === 'unity'}
        <span>🎮</span>
      {:else if tool.icon === 'blender'}
        <span style="color:#f5792a">🧊</span>
      {:else if tool.icon === 'python'}
        <span style="color:#3776ab">🐍</span>
      {:else if tool.icon === 'git'}
        <span style="color:#f05032">🔀</span>
      {:else if tool.icon === 'docker'}
        <span style="color:#2496ed">🐳</span>
      {:else if tool.icon === 'ollama'}
        <span>🦙</span>
      {:else if tool.icon === 'nvidia'}
        <span style="color:#76b900">🖥️</span>
      {:else}
        <span>📦</span>
      {/if}
    </div>
    <div class="card-info">
      <div class="tool-name-row">
        <span class="tool-name">{name}</span>
        {#if tool.isFree}
          <span class="badge badge-free">{ui['free'] || '免费'}</span>
        {:else}
          <span class="badge badge-paid">{ui['paid'] || '付费'}</span>
        {/if}
      </div>
      <p class="tool-desc">{description}</p>
    </div>
  </div>

  <div class="card-meta">
    <span class="meta-item">{ui['version'] || '版本'}: {tool.version}</span>
    <span class="meta-sep">·</span>
    <span class="meta-item">{ui['size'] || '大小'}: {tool.size}</span>
    {#if tool.homepage}
      <span class="meta-sep">·</span>
      <button class="meta-link" onclick={openHomepage}>{ui['homepage'] || '官网'} ↗</button>
    {/if}
  </div>

  {#if tool.tags.length > 0}
    <div class="card-tags">
      {#each tool.tags as tag}
        <span class="mini-tag">{tag}</span>
      {/each}
    </div>
  {/if}

  {#if installing && progress}
    <div class="progress-section">
      <div class="progress-bar">
        <div
          class="progress-fill"
          style="width: {Math.max(0, progress.percent)}%"
        ></div>
      </div>
      <span class="progress-text">{progress.message}</span>
    </div>
  {/if}

  {#if error && progress}
    <div class="error-msg">{progress.message}</div>
  {/if}

  <div class="card-actions">
    {#if done && progress}
      <span class="action-done">✓ {ui['installed'] || '已安装'}</span>
    {:else if installing}
      <button class="action-btn action-installing" disabled>
        <span class="btn-spinner"></span>
        {progress?.status === 'downloading' ? (ui['downloading'] || '下载中...') : (ui['installing'] || '安装中...')}
      </button>
    {:else if tool.installed}
      <button class="action-btn action-uninstall" onclick={onUninstall}>
        {ui['uninstall'] || '卸载'}
      </button>
    {:else}
      <button class="action-btn action-install" onclick={onInstall}>
        {tool.installType === 'winget' ? (ui['install'] || '安装') : (ui['download'] || '下载')}
      </button>
    {/if}
  </div>
</div>

<style>
  .tool-card {
    background: #161822;
    border: 1px solid #2d3148;
    border-radius: 12px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 10px;
    transition: border-color 0.15s, box-shadow 0.15s;
  }

  .tool-card:hover {
    border-color: #3d4458;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.2);
  }

  .tool-card.installing {
    border-color: #58a6ff33;
  }

  .card-header {
    display: flex;
    gap: 12px;
  }

  .tool-icon {
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24px;
    background: #1c2030;
    border-radius: 10px;
    flex-shrink: 0;
  }

  .card-info {
    flex: 1;
    min-width: 0;
  }

  .tool-name-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .tool-name {
    font-size: 15px;
    font-weight: 600;
    color: #e1e4e8;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .badge {
    font-size: 10px;
    padding: 2px 6px;
    border-radius: 4px;
    font-weight: 600;
    flex-shrink: 0;
  }

  .badge-free {
    background: #0d3321;
    color: #3fb950;
  }

  .badge-paid {
    background: #3d1f00;
    color: #d29922;
  }

  .tool-desc {
    font-size: 12px;
    color: #8b949e;
    margin-top: 4px;
    line-height: 1.4;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .card-meta {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 11px;
    color: #6e7681;
  }

  .meta-sep {
    color: #2d3148;
  }

  .meta-link {
    background: none;
    border: none;
    color: #58a6ff;
    font-size: 11px;
    cursor: pointer;
    padding: 0;
    font-family: inherit;
  }

  .meta-link:hover {
    text-decoration: underline;
  }

  .card-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }

  .mini-tag {
    font-size: 10px;
    padding: 2px 6px;
    background: #1c2030;
    border-radius: 4px;
    color: #8b949e;
  }

  .progress-section {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .progress-bar {
    height: 4px;
    background: #2d3148;
    border-radius: 2px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    background: linear-gradient(90deg, #58a6ff, #3fb950);
    border-radius: 2px;
    transition: width 0.3s ease;
  }

  .progress-text {
    font-size: 11px;
    color: #8b949e;
  }

  .error-msg {
    font-size: 11px;
    color: #f85149;
    padding: 4px 8px;
    background: #3d1214;
    border-radius: 4px;
  }

  .card-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: auto;
  }

  .action-btn {
    padding: 6px 16px;
    border: none;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s;
    font-family: inherit;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .action-install {
    background: #238636;
    color: #fff;
  }

  .action-install:hover {
    background: #2ea043;
  }

  .action-installing {
    background: #1f2940;
    color: #58a6ff;
    cursor: not-allowed;
  }

  .action-uninstall {
    background: #1c2030;
    color: #8b949e;
    border: 1px solid #2d3148;
  }

  .action-uninstall:hover {
    background: #3d1214;
    color: #f85149;
    border-color: #f85149;
  }

  .action-done {
    color: #3fb950;
    font-size: 13px;
    font-weight: 500;
  }

  .btn-spinner {
    width: 12px;
    height: 12px;
    border: 2px solid transparent;
    border-top-color: #58a6ff;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
    display: inline-block;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
