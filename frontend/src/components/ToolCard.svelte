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

  const iconMap: Record<string, { emoji: string; color: string }> = {
    vscode:   { emoji: '⟨/⟩', color: '#007ACC' },
    unity:    { emoji: '◆', color: '#FFFFFF' },
    blender:  { emoji: '△', color: '#F5792A' },
    python:   { emoji: 'λ', color: '#3776AB' },
    git:      { emoji: '⑂', color: '#F05032' },
    docker:   { emoji: '▣', color: '#2496ED' },
    ollama:   { emoji: '◎', color: '#FFFFFF' },
    nvidia:   { emoji: '◈', color: '#76B900' },
    rider:    { emoji: '⟐', color: '#DD1111' },
    godot:    { emoji: '◇', color: '#478CBF' },
    intellij: { emoji: '⊡', color: '#FE2857' },
    pycharm:  { emoji: '⊞', color: '#1ECE71' },
  };

  let iconInfo = $derived(iconMap[tool.icon] || { emoji: '□', color: '#8B949E' });

  function openHomepage() {
    if (tool.homepage) window.open(tool.homepage, '_blank');
  }
</script>

<div class="card" class:installing class:error-state={error}>
  <!-- Top glow for installing state -->
  {#if installing}
    <div class="card-glow"></div>
  {/if}

  <div class="card-top">
    <div class="icon-wrap" style="--icon-color: {iconInfo.color}">
      <span class="icon-char">{iconInfo.emoji}</span>
    </div>
    <div class="card-body">
      <div class="name-row">
        <span class="tool-name">{name}</span>
        <span class="badge" class:free={tool.isFree} class:paid={!tool.isFree}>
          {tool.isFree ? (ui['free'] || 'FREE') : (ui['paid'] || 'PRO')}
        </span>
      </div>
      <p class="tool-desc">{description}</p>
    </div>
  </div>

  <div class="meta-row">
    <span class="meta">{tool.version}</span>
    <span class="meta-dot"></span>
    <span class="meta">{tool.size}</span>
    {#if tool.homepage}
      <span class="meta-dot"></span>
      <button class="meta-link" onclick={openHomepage}>
        Homepage
        <svg viewBox="0 0 24 24" width="10" height="10" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M7 17L17 7M7 7h10v10"/></svg>
      </button>
    {/if}
  </div>

  {#if tool.tags.length > 0}
    <div class="tags-row">
      {#each tool.tags as tag}
        <span class="tag">{tag}</span>
      {/each}
    </div>
  {/if}

  {#if installing && progress}
    <div class="progress-wrap">
      <div class="progress-track">
        <div class="progress-bar" style="width: {Math.max(0, progress.percent)}%"></div>
      </div>
      <span class="progress-label">{progress.message}</span>
    </div>
  {/if}

  {#if error && progress}
    <div class="error-bar">{progress.message}</div>
  {/if}

  <div class="card-bottom">
    {#if done && progress}
      <div class="status-done">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
        {ui['installed'] || 'Installed'}
      </div>
    {:else if installing}
      <button class="btn btn-progress" disabled>
        <span class="btn-spin"></span>
        {progress?.status === 'downloading' ? (ui['downloading'] || 'Downloading') : (ui['installing'] || 'Installing')}
      </button>
    {:else if tool.installed}
      <button class="btn btn-uninstall" onclick={onUninstall}>
        {ui['uninstall'] || 'Uninstall'}
      </button>
    {:else}
      <button class="btn btn-install" onclick={onInstall}>
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M12 5v14M5 12h14"/></svg>
        {tool.installType === 'winget' ? (ui['install'] || 'Install') : (ui['download'] || 'Download')}
      </button>
    {/if}
  </div>
</div>

<style>
  .card {
    position: relative;
    background: var(--glass-2);
    backdrop-filter: var(--blur-md);
    -webkit-backdrop-filter: var(--blur-md);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-lg);
    padding: 18px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    transition: all var(--duration-normal) var(--ease-smooth);
    overflow: hidden;
  }

  .card::before {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: inherit;
    background: linear-gradient(135deg, rgba(255,255,255,0.04) 0%, transparent 60%);
    pointer-events: none;
  }

  .card:hover {
    border-color: var(--glass-border-hover);
    transform: translateY(-2px);
    box-shadow: var(--shadow-md);
  }

  .card.installing {
    border-color: rgba(10, 132, 255, 0.25);
  }

  .card.error-state {
    border-color: rgba(255, 55, 95, 0.25);
  }

  .card-glow {
    position: absolute;
    top: -40px;
    left: 50%;
    transform: translateX(-50%);
    width: 120px;
    height: 80px;
    background: radial-gradient(ellipse, rgba(10, 132, 255, 0.2) 0%, transparent 70%);
    pointer-events: none;
    animation: glow-pulse 2s ease-in-out infinite;
  }

  @keyframes glow-pulse {
    0%, 100% { opacity: 0.5; }
    50% { opacity: 1; }
  }

  .card-top {
    display: flex;
    gap: 14px;
    position: relative;
  }

  .icon-wrap {
    width: 44px;
    height: 44px;
    border-radius: var(--radius-sm);
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(255, 255, 255, 0.06);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    transition: all var(--duration-fast);
  }

  .card:hover .icon-wrap {
    background: rgba(255, 255, 255, 0.08);
    box-shadow: 0 0 16px color-mix(in srgb, var(--icon-color) 20%, transparent);
  }

  .icon-char {
    font-size: 20px;
    font-weight: 700;
    color: var(--icon-color);
    letter-spacing: -0.04em;
    font-family: 'SF Mono', 'Cascadia Code', 'Consolas', monospace;
  }

  .card-body {
    flex: 1;
    min-width: 0;
  }

  .name-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .tool-name {
    font-size: 15px;
    font-weight: 600;
    color: var(--text-primary);
    letter-spacing: -0.02em;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .badge {
    font-size: 9px;
    font-weight: 700;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    padding: 2px 7px;
    border-radius: var(--radius-full);
    flex-shrink: 0;
  }

  .badge.free {
    background: rgba(48, 209, 88, 0.12);
    color: var(--accent-green);
  }

  .badge.paid {
    background: rgba(255, 159, 10, 0.12);
    color: var(--accent-orange);
  }

  .tool-desc {
    font-size: 12.5px;
    color: var(--text-secondary);
    margin-top: 4px;
    line-height: 1.45;
    display: -webkit-box;
    line-clamp: 2;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .meta-row {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    color: var(--text-tertiary);
  }

  .meta-dot {
    width: 3px;
    height: 3px;
    border-radius: 50%;
    background: var(--text-tertiary);
    opacity: 0.5;
  }

  .meta-link {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    background: none;
    border: none;
    color: var(--accent-blue);
    font-size: 11px;
    cursor: pointer;
    padding: 0;
    font-family: inherit;
    opacity: 0.8;
    transition: opacity var(--duration-fast);
  }

  .meta-link:hover { opacity: 1; }

  .tags-row {
    display: flex;
    flex-wrap: wrap;
    gap: 5px;
  }

  .tag {
    font-size: 10px;
    font-weight: 500;
    padding: 3px 8px;
    border-radius: var(--radius-full);
    background: rgba(255, 255, 255, 0.05);
    color: var(--text-tertiary);
    letter-spacing: 0.01em;
  }

  .progress-wrap {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .progress-track {
    height: 3px;
    background: rgba(255, 255, 255, 0.06);
    border-radius: var(--radius-full);
    overflow: hidden;
  }

  .progress-bar {
    height: 100%;
    background: linear-gradient(90deg, var(--accent-blue), var(--accent-teal));
    border-radius: var(--radius-full);
    transition: width 0.4s var(--ease-smooth);
    box-shadow: 0 0 8px rgba(10, 132, 255, 0.4);
  }

  .progress-label {
    font-size: 11px;
    color: var(--text-secondary);
  }

  .error-bar {
    font-size: 11px;
    color: var(--accent-pink);
    padding: 6px 10px;
    background: rgba(255, 55, 95, 0.08);
    border: 1px solid rgba(255, 55, 95, 0.15);
    border-radius: var(--radius-sm);
  }

  .card-bottom {
    display: flex;
    align-items: center;
    margin-top: auto;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 7px 18px;
    border: none;
    border-radius: var(--radius-full);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    transition: all var(--duration-fast) var(--ease-smooth);
    font-family: inherit;
    letter-spacing: -0.01em;
  }

  .btn:active { transform: scale(0.96); }

  .btn-install {
    background: linear-gradient(135deg, var(--accent-blue), #0070E0);
    color: #fff;
    box-shadow: 0 2px 12px rgba(10, 132, 255, 0.3);
  }

  .btn-install:hover {
    box-shadow: 0 4px 20px rgba(10, 132, 255, 0.45);
    filter: brightness(1.1);
  }

  .btn-progress {
    background: rgba(10, 132, 255, 0.12);
    color: var(--accent-blue);
    cursor: not-allowed;
    box-shadow: none;
  }

  .btn-uninstall {
    background: rgba(255, 255, 255, 0.06);
    color: var(--text-secondary);
    border: 1px solid var(--glass-border);
  }

  .btn-uninstall:hover {
    background: rgba(255, 55, 95, 0.1);
    border-color: rgba(255, 55, 95, 0.25);
    color: var(--accent-pink);
  }

  .status-done {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    font-weight: 600;
    color: var(--accent-green);
  }

  .btn-spin {
    width: 12px;
    height: 12px;
    border: 2px solid transparent;
    border-top-color: var(--accent-blue);
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
