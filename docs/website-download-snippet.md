# Website Download Page Snippet

Add this to your existing website wherever you want the download section. It fetches the latest release from GitHub and renders per-platform download buttons automatically — no manual link updates needed when you publish a new version.

Replace `YOUR_GITHUB_USERNAME/YOUR_REPO_NAME` with your actual repository path in the two places it appears.

---

## HTML + JS snippet

```html
<section id="download">
  <h2>Download Genetica Resolutio</h2>
  <p>Free and open source. Your DNA never leaves your computer.</p>

  <div id="download-buttons" style="display:flex; gap:12px; flex-wrap:wrap; margin:24px 0;">
    <span id="download-loading">Loading latest release…</span>
  </div>

  <p id="download-version" style="font-size:0.85em; opacity:0.6;"></p>

  <details style="margin-top:16px; font-size:0.85em; opacity:0.7;">
    <summary>First launch security prompts</summary>
    <p><strong>macOS:</strong> Right-click the app → Open, then click Open. Or: System Settings → Privacy &amp; Security → Open Anyway. This appears because the app is unsigned open-source software.</p>
    <p><strong>Windows:</strong> Click "More info" → "Run anyway" in the SmartScreen prompt. Same reason.</p>
    <p><strong>Linux:</strong> Run <code>chmod +x genetica-resolutio-linux-amd64</code> before launching.</p>
  </details>
</section>

<script>
(function() {
  const REPO = 'YOUR_GITHUB_USERNAME/YOUR_REPO_NAME';
  const API  = `https://api.github.com/repos/${REPO}/releases/latest`;

  const PLATFORMS = [
    { label: 'Linux (x86-64)',        suffix: 'linux-amd64.tar.gz',   icon: '🐧' },
    { label: 'macOS (Intel)',          suffix: 'macos-intel.zip',      icon: '🍎' },
    { label: 'macOS (Apple Silicon)',  suffix: 'macos-arm64.zip',      icon: '🍎' },
    { label: 'Windows',               suffix: 'windows-amd64.zip',    icon: '🪟' },
  ];

  fetch(API)
    .then(r => r.json())
    .then(release => {
      const assets = release.assets || [];
      const container = document.getElementById('download-buttons');
      container.innerHTML = '';

      PLATFORMS.forEach(p => {
        const asset = assets.find(a => a.name.endsWith(p.suffix));
        if (!asset) return;
        const a = document.createElement('a');
        a.href = asset.browser_download_url;
        a.textContent = `${p.icon} ${p.label}`;
        a.style.cssText = 'display:inline-block; padding:12px 20px; background:#1a1a1a; color:#fff; border:1px solid #333; border-radius:6px; text-decoration:none; font-family:monospace;';
        container.appendChild(a);
      });

      // "View all releases" fallback link
      const all = document.createElement('a');
      all.href = `https://github.com/${REPO}/releases`;
      all.textContent = 'All releases →';
      all.style.cssText = 'display:inline-block; padding:12px 20px; color:#666; text-decoration:none; align-self:center;';
      container.appendChild(all);

      const checksums = assets.find(a => a.name === 'SHA256SUMS.txt');
      const versionEl = document.getElementById('download-version');
      versionEl.innerHTML = `Version ${release.tag_name}`;
      if (checksums) {
        versionEl.innerHTML += ` · <a href="${checksums.browser_download_url}">SHA256 checksums</a>`;
      }
      versionEl.innerHTML += ` · <a href="https://github.com/${REPO}">Source code</a>`;
    })
    .catch(() => {
      document.getElementById('download-buttons').innerHTML =
        `<a href="https://github.com/${REPO}/releases/latest" style="padding:12px 20px; background:#1a1a1a; color:#fff; border-radius:6px; text-decoration:none;">View latest release on GitHub →</a>`;
    });
})();
</script>
```

---

## Release process

Once your code is in a GitHub repository:

```bash
# Tag a release — this triggers the build automatically
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions builds Linux, macOS (Intel), macOS (ARM), and Windows in parallel (~10–15 minutes), then creates a Release with all four binaries attached. The download buttons on your site update automatically.

---

## Directing users who want to verify or build themselves

Link to `https://github.com/YOUR_USERNAME/YOUR_REPO` and the `docs/building-from-source.md` page. For a privacy-sensitive app, having the source publicly auditable is itself a feature worth mentioning on the download page.
