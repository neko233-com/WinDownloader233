<script lang="ts">
  interface Props {
    query: string;
    allTags: string[];
    activeTags: string[];
    ui: Record<string, string>;
    loading: boolean;
    statusMessage: string;
    onSearch: (q: string) => void;
    onToggleTag: (tag: string) => void;
    onRefresh: () => void;
  }

  let { query, allTags, activeTags, ui, loading, statusMessage, onSearch, onToggleTag, onRefresh }: Props = $props();

  let inputValue = $state('');
  let searchTimeout: ReturnType<typeof setTimeout>;
  let focused = $state(false);

  function handleInput() {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => onSearch(inputValue), 300);
  }

  function clearSearch() {
    inputValue = '';
    onSearch('');
  }
</script>

<header class="header">
  <div class="header-top">
    <div class="search-box" class:focused>
      <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="11" cy="11" r="8"/>
        <path d="m21 21-4.35-4.35"/>
      </svg>
      <input
        type="text"
        class="search-input"
        placeholder={ui['search'] || '搜索工具...'}
        bind:value={inputValue}
        oninput={handleInput}
        onfocus={() => focused = true}
        onblur={() => focused = false}
      />
      {#if inputValue}
        <button class="search-clear" onclick={clearSearch} aria-label="Clear search">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M12 2C6.47 2 2 6.47 2 12s4.47 10 10 10 10-4.47 10-10S17.53 2 12 2zm5 13.59L15.59 17 12 13.41 8.41 17 7 15.59 10.59 12 7 8.41 8.41 7 12 10.59 15.59 7 17 8.41 13.41 12 17 15.59z"/></svg>
        </button>
      {/if}
    </div>

    <button class="refresh-btn" onclick={onRefresh} title={ui['refresh'] || '刷新'} disabled={loading}>
      {#if loading}
        <span class="spinner"></span>
      {:else}
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.3"/>
        </svg>
      {/if}
    </button>
  </div>

  {#if statusMessage}
    <div class="status-toast">
      <span class="status-dot"></span>
      {statusMessage}
    </div>
  {/if}

  {#if allTags.length > 0}
    <div class="tag-bar">
      {#each allTags.slice(0, 24) as tag}
        <button
          class="tag-pill"
          class:active={activeTags.includes(tag)}
          onclick={() => onToggleTag(tag)}
        >
          {tag}
        </button>
      {/each}
    </div>
  {/if}
</header>

<style>
  .header {
    padding: 18px 28px 14px;
    flex-shrink: 0;
  }

  .header-top {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .search-box {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 8px;
    background: var(--glass-2);
    backdrop-filter: var(--blur-md);
    -webkit-backdrop-filter: var(--blur-md);
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-md);
    padding: 0 14px;
    transition: all var(--duration-normal) var(--ease-smooth);
  }

  .search-box.focused {
    border-color: rgba(10, 132, 255, 0.4);
    background: var(--glass-1);
    box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.1), var(--shadow-sm);
  }

  .search-icon {
    width: 16px;
    height: 16px;
    color: var(--text-tertiary);
    flex-shrink: 0;
    transition: color var(--duration-fast);
  }

  .search-box.focused .search-icon {
    color: var(--accent-blue);
  }

  .search-input {
    flex: 1;
    background: transparent;
    border: none;
    outline: none;
    color: var(--text-primary);
    font-size: 14px;
    font-weight: 400;
    padding: 11px 0;
    letter-spacing: -0.01em;
  }

  .search-input::placeholder {
    color: var(--text-tertiary);
  }

  .search-clear {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    border: none;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.12);
    color: var(--text-secondary);
    cursor: pointer;
    transition: all var(--duration-fast);
    flex-shrink: 0;
  }

  .search-clear:hover {
    background: rgba(255, 255, 255, 0.2);
    color: var(--text-primary);
  }

  .refresh-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-sm);
    background: var(--glass-2);
    backdrop-filter: var(--blur-md);
    -webkit-backdrop-filter: var(--blur-md);
    color: var(--text-secondary);
    cursor: pointer;
    transition: all var(--duration-fast) var(--ease-smooth);
    flex-shrink: 0;
  }

  .refresh-btn:hover:not(:disabled) {
    background: var(--glass-1);
    color: var(--text-primary);
    border-color: var(--glass-border-hover);
  }

  .refresh-btn:active:not(:disabled) {
    transform: scale(0.92);
  }

  .refresh-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .spinner {
    width: 16px;
    height: 16px;
    border: 2px solid transparent;
    border-top-color: var(--accent-blue);
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .status-toast {
    margin-top: 10px;
    padding: 8px 14px;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    background: var(--glass-2);
    backdrop-filter: var(--blur-md);
    -webkit-backdrop-filter: var(--blur-md);
    border: 1px solid rgba(10, 132, 255, 0.15);
    border-radius: var(--radius-full);
    color: var(--accent-teal);
    font-size: 12px;
    font-weight: 500;
    animation: slideDown 0.3s var(--ease-spring);
  }

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--accent-teal);
    box-shadow: 0 0 8px var(--accent-teal);
  }

  @keyframes slideDown {
    from { opacity: 0; transform: translateY(-8px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .tag-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 12px;
  }

  .tag-pill {
    padding: 5px 12px;
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-full);
    background: var(--glass-3);
    backdrop-filter: var(--blur-sm);
    -webkit-backdrop-filter: var(--blur-sm);
    color: var(--text-secondary);
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: all var(--duration-fast) var(--ease-smooth);
    letter-spacing: 0.01em;
  }

  .tag-pill:hover {
    border-color: var(--glass-border-hover);
    color: var(--text-primary);
    background: var(--glass-2);
  }

  .tag-pill:active {
    transform: scale(0.95);
  }

  .tag-pill.active {
    background: rgba(10, 132, 255, 0.15);
    border-color: rgba(10, 132, 255, 0.3);
    color: var(--accent-blue);
    box-shadow: 0 0 12px rgba(10, 132, 255, 0.1);
  }
</style>
