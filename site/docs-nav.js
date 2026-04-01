/**
 * Shared documentation sidebar — injected into every doc page.
 * Highlights the current page automatically.
 */
(function () {
  const nav = [
    { href: 'index.html', label: 'Overview', icon: '&#x1F4D6;' },
    { href: 'deployment.html', label: 'Deployment', icon: '&#x1F680;' },
    { href: 'configuration.html', label: 'Configuration', icon: '&#x2699;' },
    { href: 'modules.html', label: 'Modules', icon: '&#x1F9E9;' },
    { href: 'api.html', label: 'API Reference', icon: '&#x1F4E1;' },
    { href: 'translating.html', label: 'Translating', icon: '&#x1F310;' },
  ];

  const current = location.pathname.split('/').pop() || 'index.html';

  const sidebar = document.getElementById('docs-sidebar');
  if (!sidebar) return;

  let html = '<div class="sidebar-header">';
  html += '<a href="../" class="sidebar-brand">';
  html += '<img src="../img/logo.png" alt="Nidus" width="28" height="28">';
  html += '<span>Nidus</span></a>';
  html += '<button class="sidebar-close" aria-label="Close menu" onclick="document.getElementById(\'docs-sidebar\').classList.remove(\'open\')">';
  html += '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6L6 18M6 6l12 12"/></svg>';
  html += '</button></div>';
  html += '<div class="sidebar-label">Documentation</div>';
  html += '<ul class="sidebar-nav">';

  nav.forEach(function (item) {
    const active = current === item.href ? ' class="active"' : '';
    html += '<li><a href="' + item.href + '"' + active + '>';
    html += '<span class="sidebar-icon">' + item.icon + '</span>';
    html += item.label + '</a></li>';
  });

  html += '</ul>';
  html += '<div class="sidebar-footer">';
  html += '<a href="https://github.com/tdebuilt/Nidus-Dashboard" target="_blank" rel="noopener">GitHub</a>';
  html += '<a href="https://github.com/tdebuilt/Nidus-Dashboard/releases" target="_blank" rel="noopener">Releases</a>';
  html += '</div>';

  sidebar.innerHTML = html;
})();
