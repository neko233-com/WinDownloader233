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
    <span class="brand-icon">📦</span>
    <span class="brand-text">UniGet<span class="brand-accent">233</span></span>
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
      </button>
    {/each}
  </div>

  <div class="sidebar-footer">
    <button class="nav-item settings-btn" onclick={onSettings}>
      <span class="nav-icon">⚙️</span>
      <span class="nav-label">设置</span>
    </button>
  </div>
</nav>

<style>
  .sidebar {
    width: 200px;
    min-width: 200px;
    height: 100vh;
    background: #161822;
    border-right: 1px solid #2d3148;
    display: flex;
    flex-direction: column;
    user-select: none;
  }

  .sidebar-brand {
    padding: 20px 16px;
    display: flex;
    align-items: center;
    gap: 10px;
    border-bottom: 1px solid #2d3148;
  }

  .brand-icon {
    font-size: 24px;
  }

  .brand-text {
    font-size: 18px;
    font-weight: 700;
    color: #c9d1d9;
    letter-spacing: -0.5px;
  }

  .brand-accent {
    color: #58a6ff;
  }

  .sidebar-nav {
    flex: 1;
    padding: 12px 8px;
    display: flex;
    flex-direction: column;
    gap: 2px;
    overflow-y: auto;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 12px;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: #8b949e;
    font-size: 14px;
    cursor: pointer;
    transition: all 0.15s ease;
    text-align: left;
    width: 100%;
  }

  .nav-item:hover {
    background: #1c2030;
    color: #c9d1d9;
  }

  .nav-item.active {
    background: #1f2940;
    color: #58a6ff;
    font-weight: 600;
  }

  .nav-icon {
    font-size: 18px;
    width: 24px;
    text-align: center;
    flex-shrink: 0;
  }

  .nav-label {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .sidebar-footer {
    padding: 8px;
    border-top: 1px solid #2d3148;
  }

  .settings-btn {
    color: #6e7681;
  }

  .settings-btn:hover {
    color: #c9d1d9;
  }
</style>
