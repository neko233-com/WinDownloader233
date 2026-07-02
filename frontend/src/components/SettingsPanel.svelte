<script lang="ts">
  import { AppService } from '../../bindings/github.com/neko233/WinDownloader233';

  interface Props {
    lang: string;
    ui: Record<string, string>;
    onLanguageChange: (lang: string) => void;
    onClose: () => void;
  }

  let { lang, ui, onLanguageChange, onClose }: Props = $props();

  let useProxy = $state(false);
  let githubProxy = $state('https://ghfast.top/');
  let httpProxy = $state('');
  let visible = $state(false);

  async function loadConfig() {
    try {
      const cfg = await AppService.GetMirrorConfig();
      useProxy = cfg.useProxy;
      githubProxy = cfg.githubProxy;
      httpProxy = cfg.httpProxy;
    } catch {}
  }

  async function saveMirrorConfig() {
    try {
      await AppService.SetMirrorConfig({ githubProxy, downloadCdn: githubProxy, useProxy, httpProxy });
    } catch {}
  }

  function selectProxy(url: string) {
    if (url) { githubProxy = url; saveMirrorConfig(); }
  }

  let proxyOptions = $derived([
    { name: 'ghfast.top', url: 'https://ghfast.top/', desc: lang === 'zh' ? '高速稳定' : 'Fast & stable' },
    { name: 'gh-proxy.com', url: 'https://gh-proxy.com/', desc: lang === 'zh' ? '免费公共代理' : 'Free public proxy' },
    { name: 'mirror.ghproxy.com', url: 'https://mirror.ghproxy.com/', desc: 'GitHub mirror' },
  ]);

  loadConfig();
  // Animate in
  requestAnimationFrame(() => { visible = true; });

  function handleClose() {
    visible = false;
    setTimeout(onClose, 250);
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="overlay" class:visible onclick={handleClose} onkeydown={() => {}}>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="sheet" class:visible onclick={(e) => e.stopPropagation()} onkeydown={() => {}}>
    <!-- Sheet handle -->
    <div class="sheet-handle">
      <div class="handle-bar"></div>
    </div>

    <div class="sheet-header">
      <h2 class="sheet-title">{ui['settings'] || 'Settings'}</h2>
      <button class="close-btn" onclick={handleClose} aria-label="Close">
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M18 6L6 18M6 6l12 12"/></svg>
      </button>
    </div>

    <div class="sheet-body">
      <!-- Language -->
      <section class="section">
        <div class="section-header">
          <div class="section-icon lang-icon">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="10"/><path d="M2 12h20M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
          </div>
          <h3 class="section-title">{ui['language'] || 'Language'}</h3>
        </div>
        <div class="seg-control">
          <button class="seg-btn" class:active={lang === 'zh'} onclick={() => onLanguageChange('zh')}>
            中文
          </button>
          <button class="seg-btn" class:active={lang === 'en'} onclick={() => onLanguageChange('en')}>
            English
          </button>
        </div>
      </section>

      <!-- Divider -->
      <div class="divider"></div>

      <!-- Mirror -->
      <section class="section">
        <div class="section-header">
          <div class="section-icon proxy-icon">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
          </div>
          <h3 class="section-title">{ui['mirror'] || 'Mirror Proxy'}</h3>
        </div>
        <p class="section-desc">
          {lang === 'zh' ? 'GitHub 镜像代理加速下载' : 'GitHub mirror proxy for faster downloads'}
        </p>

        <div class="toggle-row" onclick={() => { useProxy = !useProxy; saveMirrorConfig(); }} onkeydown={() => {}} role="switch" aria-checked={useProxy} tabindex="0">
          <span class="toggle-label">{lang === 'zh' ? '启用代理' : 'Enable Proxy'}</span>
          <div class="toggle" class:on={useProxy}>
            <div class="toggle-knob"></div>
          </div>
        </div>

        {#if useProxy}
          <div class="proxy-section">
            {#each proxyOptions as opt}
              <button class="proxy-option" class:active={githubProxy === opt.url} onclick={() => selectProxy(opt.url)}>
                <div class="proxy-opt-left">
                  <span class="proxy-opt-name">{opt.name}</span>
                  <span class="proxy-opt-desc">{opt.desc}</span>
                </div>
                {#if githubProxy === opt.url}
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><polyline points="20 6 9 17 4 12"/></svg>
                {/if}
              </button>
            {/each}

            <div class="input-group">
              <label class="input-label" for="custom-proxy">{lang === 'zh' ? '自定义代理' : 'Custom Proxy'}</label>
              <input id="custom-proxy" type="text" class="glass-input" placeholder="https://your-proxy.com/" bind:value={githubProxy} onchange={saveMirrorConfig} />
            </div>

            <div class="input-group">
              <label class="input-label" for="http-proxy">HTTP Proxy</label>
              <input id="http-proxy" type="text" class="glass-input" placeholder="http://127.0.0.1:7890" bind:value={httpProxy} onchange={saveMirrorConfig} />
              <span class="input-hint">{lang === 'zh' ? '本地代理（Clash / V2Ray）' : 'Local proxy (Clash / V2Ray)'}</span>
            </div>
          </div>
        {/if}
      </section>

      <!-- Divider -->
      <div class="divider"></div>

      <!-- Registry Info -->
      <section class="section">
        <div class="section-header">
          <div class="section-icon info-icon">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4M12 8h.01"/></svg>
          </div>
          <h3 class="section-title">{lang === 'zh' ? '数据源' : 'Registry'}</h3>
        </div>
        <p class="section-desc">
          {lang === 'zh' ? '工具列表由 GitHub JSON 驱动，本地 + 远程时间戳对比取最新，无需更新应用。' : 'Tool list driven by GitHub JSON. Local + remote compared by timestamp, newest wins.'}
        </p>
      </section>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    z-index: 1000;
    background: rgba(15, 23, 42, 0);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background var(--duration-normal) var(--ease-smooth);
  }

  .overlay.visible {
    background: rgba(15, 23, 42, 0.24);
  }

  .sheet {
    width: 560px;
    max-width: 92vw;
    max-height: 86vh;
    background: #fff;
    border: 1px solid #cbd5e1;
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    box-shadow: 0 24px 70px rgba(15, 23, 42, 0.22);
    opacity: 0;
    transform: translateY(10px);
    transition: opacity var(--duration-normal) var(--ease-smooth), transform var(--duration-normal) var(--ease-smooth);
    overflow: hidden;
  }

  .sheet.visible {
    opacity: 1;
    transform: translateY(0);
  }

  .sheet-handle {
    display: none;
  }

  .handle-bar {
    width: 36px;
    height: 4px;
    border-radius: var(--radius-full);
    background: rgba(255, 255, 255, 0.15);
  }

  .sheet-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 18px 20px;
    border-bottom: 1px solid #e2e8f0;
    background: #f8fafc;
  }

  .sheet-title {
    font-size: 19px;
    font-weight: 700;
    letter-spacing: 0;
    color: var(--text-primary);
  }

  .close-btn {
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    border-radius: 8px;
    background: #fff;
    border: 1px solid #d8e0ea;
    color: var(--text-secondary);
    cursor: pointer;
    transition: all var(--duration-fast);
  }

  .close-btn:hover {
    background: #f1f5f9;
    color: var(--text-primary);
  }

  .sheet-body {
    padding: 16px 20px 20px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 12px;
    background: #fff;
  }

  .section {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 14px;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    background: #fff;
  }

  .section-header {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .section-icon {
    width: 32px;
    height: 32px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .lang-icon {
    background: #e8f3ff;
    color: var(--accent-blue);
  }

  .proxy-icon {
    background: #fff7ed;
    color: #ea580c;
  }

  .info-icon {
    background: #f5f3ff;
    color: #7c3aed;
  }

  .section-title {
    font-size: 15px;
    font-weight: 700;
    color: var(--text-primary);
    letter-spacing: 0;
  }

  .section-desc {
    font-size: 13px;
    color: var(--text-secondary);
    line-height: 1.5;
  }

  .divider {
    display: none;
  }

  /* Segmented control */
  .seg-control {
    display: flex;
    background: #f1f5f9;
    border: 1px solid #d8e0ea;
    border-radius: 8px;
    padding: 2px;
    gap: 2px;
  }

  .seg-btn {
    flex: 1;
    padding: 8px 0;
    border: none;
    border-radius: 6px;
    background: transparent;
    color: var(--text-secondary);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all var(--duration-fast) var(--ease-smooth);
    font-family: inherit;
  }

  .seg-btn:hover { color: var(--text-primary); }

  .seg-btn.active {
    background: #fff;
    color: var(--text-primary);
    font-weight: 700;
    box-shadow: var(--shadow-sm);
  }

  /* Toggle switch */
  .toggle-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    cursor: pointer;
  }

  .toggle-label {
    font-size: 14px;
    color: var(--text-primary);
    font-weight: 500;
  }

  .toggle {
    position: relative;
    width: 44px;
    height: 26px;
    border-radius: 13px;
    background: #cbd5e1;
    cursor: pointer;
    transition: background var(--duration-normal) var(--ease-smooth);
  }

  .toggle.on {
    background: #0A84FF;
  }

  .toggle-knob {
    position: absolute;
    top: 3px;
    left: 3px;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 1px 3px rgba(15, 23, 42, 0.25);
    transition: transform var(--duration-normal) var(--ease-spring);
  }

  .toggle.on .toggle-knob {
    transform: translateX(18px);
  }

  /* Proxy options */
  .proxy-section {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 8px 0 0;
  }

  .proxy-option {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    border: 1px solid #d8e0ea;
    border-radius: 8px;
    background: #fff;
    cursor: pointer;
    transition: all var(--duration-fast);
    text-align: left;
    font-family: inherit;
    color: var(--text-primary);
    width: 100%;
  }

  .proxy-option:hover {
    background: #f8fafc;
  }

  .proxy-option.active {
    background: #e8f3ff;
    border-color: #0A84FF;
    color: var(--accent-blue);
  }

  .proxy-opt-left {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .proxy-opt-name {
    font-size: 13px;
    font-weight: 500;
  }

  .proxy-opt-desc {
    font-size: 11px;
    color: var(--text-tertiary);
  }

  /* Inputs */
  .input-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-top: 4px;
  }

  .input-label {
    font-size: 12px;
    color: var(--text-tertiary);
    font-weight: 500;
  }

  .glass-input {
    padding: 10px 12px;
    background: #fff;
    border: 1px solid #cbd5e1;
    border-radius: 8px;
    color: var(--text-primary);
    font-size: 13px;
    font-family: 'SF Mono', 'Cascadia Code', 'Consolas', monospace;
    outline: none;
    transition: border-color var(--duration-fast);
  }

  .glass-input:focus {
    border-color: #0A84FF;
    box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.14);
  }

  .glass-input::placeholder {
    color: var(--text-tertiary);
  }

  .input-hint {
    font-size: 11px;
    color: var(--text-tertiary);
    line-height: 1.4;
  }
</style>
