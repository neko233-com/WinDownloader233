<script lang="ts">
  import type { Category } from '../lib/types';

  interface Props {
    category: Category;
    categoryNames: Record<string, string>;
    CATEGORY_ICONS: Record<Category, string>;
    ALL_CATEGORIES: Category[];
    onSelect: (cat: Category) => void;
    onSettings: () => void;
  }

  let { category, categoryNames, CATEGORY_ICONS, ALL_CATEGORIES, onSelect, onSettings }: Props = $props();
</script>

<nav class="sidebar">
  <div class="sidebar-brand">
    <div class="brand-mark">
      <span class="brand-glyph">W</span>
    </div>
    <div class="brand-text-group">
      <span class="brand-title">WinDownloader</span>
      <span class="brand-sub">233</span>
    </div>
  </div>

  <div class="sidebar-nav">
    {#each ALL_CATEGORIES as cat}
      <button
        class="nav-item"
        class:active={category === cat}
        onclick={() => onSelect(cat)}
      >
        <span class="nav-icon">{CATEGORY_ICONS[cat]}</span>
        <span class="nav-label">{categoryNames[cat] || cat}</span>
        {#if category === cat}
          <span class="nav-indicator"></span>
        {/if}
      </button>
    {/each}
  </div>

  <div class="sidebar-footer">
    <button class="nav-item settings-item" onclick={onSettings}>
      <svg class="settings-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="3"/>
        <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
      </svg>
      <span class="nav-label">Settings</span>
    </button>
  </div>
</nav>

<style>
  .sidebar {
    width: 220px;
    min-width: 220px;
    height: 100vh;
    background: var(--glass-1);
    backdrop-filter: var(--blur-lg);
    -webkit-backdrop-filter: var(--blur-lg);
    border-right: 1px solid var(--glass-border);
    display: flex;
    flex-direction: column;
    user-select: none;
    position: relative;
  }

  .sidebar::after {
    content: '';
    position: absolute;
    top: 0;
    right: 0;
    width: 1px;
    height: 100%;
    background: linear-gradient(to bottom,
      rgba(255,255,255,0.06) 0%,
      rgba(255,255,255,0.02) 50%,
      rgba(255,255,255,0.06) 100%
    );
  }

  .sidebar-brand {
    padding: 24px 18px 20px;
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .brand-mark {
    width: 36px;
    height: 36px;
    border-radius: var(--radius-sm);
    background: linear-gradient(135deg, var(--accent-blue), var(--accent-purple));
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 4px 12px rgba(10, 132, 255, 0.3);
    flex-shrink: 0;
  }

  .brand-glyph {
    font-size: 18px;
    font-weight: 700;
    color: #fff;
    letter-spacing: -0.02em;
  }

  .brand-text-group {
    display: flex;
    align-items: baseline;
    gap: 2px;
  }

  .brand-title {
    font-size: 15px;
    font-weight: 600;
    color: var(--text-primary);
    letter-spacing: -0.02em;
  }

  .brand-sub {
    font-size: 12px;
    font-weight: 500;
    color: var(--accent-blue);
    opacity: 0.8;
  }

  .sidebar-nav {
    flex: 1;
    padding: 8px 10px;
    display: flex;
    flex-direction: column;
    gap: 2px;
    overflow-y: auto;
  }

  .nav-item {
    position: relative;
    display: flex;
    align-items: center;
    gap: 11px;
    padding: 10px 12px;
    border: none;
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--text-secondary);
    font-size: 13.5px;
    font-weight: 500;
    cursor: pointer;
    transition: all var(--duration-fast) var(--ease-smooth);
    text-align: left;
    width: 100%;
    overflow: hidden;
  }

  .nav-item:hover {
    background: var(--glass-highlight);
    color: var(--text-primary);
  }

  .nav-item:active {
    transform: scale(0.97);
  }

  .nav-item.active {
    background: rgba(10, 132, 255, 0.12);
    color: var(--accent-blue);
    font-weight: 600;
  }

  .nav-item.active:hover {
    background: rgba(10, 132, 255, 0.18);
  }

  .nav-icon {
    font-size: 17px;
    width: 24px;
    text-align: center;
    flex-shrink: 0;
    line-height: 1;
  }

  .nav-label {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    flex: 1;
  }

  .nav-indicator {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    background: var(--accent-blue);
    box-shadow: 0 0 8px var(--accent-blue);
    flex-shrink: 0;
  }

  .sidebar-footer {
    padding: 10px;
    border-top: 1px solid var(--glass-border);
  }

  .settings-item {
    color: var(--text-tertiary);
  }

  .settings-item:hover {
    color: var(--text-secondary);
  }

  .settings-icon {
    width: 18px;
    height: 18px;
    flex-shrink: 0;
  }
</style>
